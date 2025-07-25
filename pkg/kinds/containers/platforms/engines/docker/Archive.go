package docker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (container *Docker) archive(buf *bytes.Buffer, hostPath, mountPoint string, permissions *v1.FileInfo) error {
	gzw := gzip.NewWriter(buf)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(hostPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(hostPath, filePath)
		if err != nil {
			return err
		}

		archivePath := filepath.Join(mountPoint, relPath)
		archivePath = strings.ReplaceAll(archivePath, "\\", "/")

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = archivePath

		if permissions != nil {
			if permissions.Owner != nil {
				header.Uid = *permissions.Owner
			}
			if permissions.Group != nil {
				header.Gid = *permissions.Group
			}
			if permissions.Permissions != nil {
				header.Mode = int64(*permissions.Permissions)
			}
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			return err
		}

		return nil
	})
}
