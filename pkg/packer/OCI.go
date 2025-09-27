package packer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/opencontainers/image-spec/specs-go"
	"io"
	"oras.land/oras-go/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	DefaultRegistry     = "registry.simplecontainer.io"
	PackageMetadataFile = "Pack.yaml"
	DefaultMediaType    = "application/vnd.simplecontainer.pack.v1+json"
)

type PackageMetadata struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

type Config struct {
	Registry  string
	Username  string
	Password  string
	Insecure  bool
	PlainHTTP bool
	Debug     bool
}

type Client struct {
	config   *Config
	registry *remote.Repository
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Registry == "" {
		config.Registry = DefaultRegistry
	}

	client := &Client{
		config: config,
	}

	return client, nil
}

func (c *Client) initRegistry(repository string) error {
	reg, err := remote.NewRepository(fmt.Sprintf("%s/%s", c.config.Registry, repository))
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	if c.config.Username != "" && c.config.Password != "" {
		reg.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(c.config.Registry, auth.Credential{
				Username: c.config.Username,
				Password: c.config.Password,
			}),
		}
	} else {
		store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
		if err != nil {
			return fmt.Errorf("failed to create credential store: %w", err)
		}

		reg.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: credentials.Credential(store),
		}
	}

	reg.PlainHTTP = c.config.PlainHTTP

	c.registry = reg
	return nil
}

func (c *Client) CreatePackage(ctx context.Context, sourceDir, outputDir string) (*PackageMetadata, error) {
	metadata, err := c.readPackageMetadata(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read package metadata: %w", err)
	}

	if err := c.validateMetadata(metadata); err != nil {
		return nil, fmt.Errorf("invalid package metadata: %w", err)
	}

	blobsDir := filepath.Join(outputDir, "blobs", "sha256")
	if err := os.MkdirAll(blobsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	layerDigest, layerSize, err := c.createLayer(sourceDir, blobsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create layer: %w", err)
	}

	configDigest, configSize, err := c.createConfig(metadata, layerDigest, blobsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	manifestDigest, err := c.createManifest(metadata, configDigest, configSize, layerDigest, layerSize, blobsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}

	if err := c.createIndex(metadata, manifestDigest, outputDir); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	if err := c.createLayout(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create layout: %w", err)
	}

	return metadata, nil
}

func (c *Client) UploadPackage(ctx context.Context, packagePath, repository, tag string) error {
	if err := c.initRegistry(repository); err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}

	store, err := file.New(packagePath)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer store.Close()

	indexPath := filepath.Join(packagePath, "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	var index v1.Index
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("failed to unmarshal index: %w", err)
	}

	if len(index.Manifests) == 0 {
		return fmt.Errorf("no manifests found in index")
	}

	manifestDescriptor := index.Manifests[0]

	_, err = oras.Copy(ctx, store, manifestDescriptor.Digest.String(), c.registry, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to push package: %w", err)
	}

	return nil
}

func (c *Client) DownloadPackage(ctx context.Context, repository, tag, outputDir string) (*PackageMetadata, error) {
	if err := c.initRegistry(repository); err != nil {
		return nil, fmt.Errorf("failed to initialize registry: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "ocipack-download-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := file.New(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create file store: %w", err)
	}
	defer store.Close()

	descriptor, err := oras.Copy(ctx, c.registry, tag, store, tag, oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to pull package: %w", err)
	}

	metadata, err := c.extractPackage(ctx, store, descriptor, outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract package: %w", err)
	}

	return metadata, nil
}

func (c *Client) readPackageMetadata(sourceDir string) (*PackageMetadata, error) {
	metadataPath := filepath.Join(sourceDir, PackageMetadataFile)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", PackageMetadataFile, err)
	}

	var metadata PackageMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", PackageMetadataFile, err)
	}

	return &metadata, nil
}

func (c *Client) validateMetadata(metadata *PackageMetadata) error {
	if metadata.Name == "" {
		return fmt.Errorf("package name is required")
	}
	if metadata.Version == "" {
		return fmt.Errorf("package version is required")
	}
	return nil
}

func (c *Client) createLayer(sourceDir, blobsDir string) (digest.Digest, int64, error) {
	tempFile, err := os.CreateTemp(blobsDir, "layer-*.tar.gz")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	file, err := os.Create(tempPath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hasher := sha256.New()
	multiWriter := io.MultiWriter(file, hasher)

	gzipWriter := gzip.NewWriter(multiWriter)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = strings.ReplaceAll(relPath, "\\", "/")

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			if _, err := io.Copy(tarWriter, srcFile); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		os.Remove(tempPath)
		return "", 0, err
	}

	tarWriter.Close()
	gzipWriter.Close()

	stat, err := file.Stat()
	if err != nil {
		os.Remove(tempPath)
		return "", 0, err
	}

	layerDigest := digest.NewDigest("sha256", hasher)
	finalPath := filepath.Join(blobsDir, layerDigest.Hex())

	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath)
		return "", 0, fmt.Errorf("failed to rename layer file: %w", err)
	}

	return layerDigest, stat.Size(), nil
}

func (c *Client) createConfig(metadata *PackageMetadata, layerDigest digest.Digest, blobsDir string) (digest.Digest, int64, error) {
	config := v1.Image{
		Created: &time.Time{},
		Author:  "ocipack",
		Config: v1.ImageConfig{
			Labels: map[string]string{
				"org.opencontainers.image.title":       metadata.Name,
				"org.opencontainers.image.version":     metadata.Version,
				"org.opencontainers.image.description": fmt.Sprintf("OCI package: %s", metadata.Name),
				"org.opencontainers.image.created":     time.Now().Format(time.RFC3339),
				"simplecontainer.pack.name":            metadata.Name,
				"simplecontainer.pack.version":         metadata.Version,
			},
		},
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{layerDigest},
		},
		History: []v1.History{
			{
				Created:   &time.Time{},
				CreatedBy: "ocipack",
				Comment:   fmt.Sprintf("Package: %s v%s", metadata.Name, metadata.Version),
			},
		},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", 0, err
	}

	configDigest := digest.FromBytes(configJSON)
	configPath := filepath.Join(blobsDir, configDigest.Hex())

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return "", 0, err
	}

	return configDigest, int64(len(configJSON)), nil
}

func (c *Client) createManifest(metadata *PackageMetadata, configDigest digest.Digest, configSize int64, layerDigest digest.Digest, layerSize int64, blobsDir string) (digest.Digest, error) {
	manifest := v1.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		MediaType: v1.MediaTypeImageManifest,
		Config: v1.Descriptor{
			MediaType: v1.MediaTypeImageConfig,
			Digest:    configDigest,
			Size:      configSize,
		},
		Layers: []v1.Descriptor{
			{
				MediaType: v1.MediaTypeImageLayerGzip,
				Digest:    layerDigest,
				Size:      layerSize,
			},
		},
		Annotations: map[string]string{
			"org.opencontainers.image.title":       metadata.Name,
			"org.opencontainers.image.version":     metadata.Version,
			"org.opencontainers.image.description": fmt.Sprintf("OCI package: %s", metadata.Name),
			"org.opencontainers.image.created":     time.Now().Format(time.RFC3339),
			"simplecontainer.pack.name":            metadata.Name,
			"simplecontainer.pack.version":         metadata.Version,
		},
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}

	manifestDigest := digest.FromBytes(manifestJSON)
	manifestPath := filepath.Join(blobsDir, manifestDigest.Hex())

	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return "", err
	}

	return manifestDigest, nil
}

func (c *Client) createIndex(metadata *PackageMetadata, manifestDigest digest.Digest, outputDir string) error {
	manifestPath := filepath.Join(outputDir, "blobs", "sha256", manifestDigest.Hex())
	stat, err := os.Stat(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to stat manifest: %w", err)
	}

	index := v1.Index{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		MediaType: v1.MediaTypeImageIndex,
		Manifests: []v1.Descriptor{
			{
				MediaType: v1.MediaTypeImageManifest,
				Digest:    manifestDigest,
				Size:      stat.Size(),
				Annotations: map[string]string{
					"org.opencontainers.image.ref.name": fmt.Sprintf("%s:%s", metadata.Name, metadata.Version),
				},
			},
		},
		Annotations: map[string]string{
			"org.opencontainers.image.title":       metadata.Name,
			"org.opencontainers.image.version":     metadata.Version,
			"org.opencontainers.image.description": fmt.Sprintf("OCI package index: %s", metadata.Name),
			"simplecontainer.pack.name":            metadata.Name,
			"simplecontainer.pack.version":         metadata.Version,
		},
	}

	indexJSON, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	indexPath := filepath.Join(outputDir, "index.json")
	return os.WriteFile(indexPath, indexJSON, 0644)
}

func (c *Client) createLayout(outputDir string) error {
	layout := v1.ImageLayout{
		Version: "1.0.0",
	}

	layoutJSON, err := json.MarshalIndent(layout, "", "  ")
	if err != nil {
		return err
	}

	layoutPath := filepath.Join(outputDir, "oci-layout")
	return os.WriteFile(layoutPath, layoutJSON, 0644)
}

func (c *Client) extractPackage(ctx context.Context, store *file.Store, descriptor v1.Descriptor, outputDir string) (*PackageMetadata, error) {
	manifestReader, err := store.Fetch(ctx, descriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer manifestReader.Close()

	manifestData, err := io.ReadAll(manifestReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	metadata := &PackageMetadata{
		Name:    manifest.Annotations["simplecontainer.pack.name"],
		Version: manifest.Annotations["simplecontainer.pack.version"],
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, layer := range manifest.Layers {
		if err := c.extractLayer(ctx, store, layer, outputDir); err != nil {
			return nil, fmt.Errorf("failed to extract layer: %w", err)
		}
	}

	return metadata, nil
}

func (c *Client) extractLayer(ctx context.Context, store *file.Store, layerDesc v1.Descriptor, outputDir string) error {
	layerReader, err := store.Fetch(ctx, layerDesc)
	if err != nil {
		return fmt.Errorf("failed to fetch layer: %w", err)
	}
	defer layerReader.Close()

	gzipReader, err := gzip.NewReader(layerReader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath := filepath.Join(outputDir, header.Name)

		if !strings.HasPrefix(targetPath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
			}

			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			file.Close()
		}
	}

	return nil
}
