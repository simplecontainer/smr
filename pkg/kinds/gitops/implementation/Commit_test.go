package implementation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/simplecontainer/smr/pkg/definitions"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// UNIT TESTS: Commit Operations
// ============================================================================

func TestNewCommit(t *testing.T) {
	commit := NewCommit()

	assert.NotNil(t, commit)
	assert.NotNil(t, commit.Format)
	assert.Nil(t, commit.Patch)
	assert.Empty(t, commit.Message)
	assert.Nil(t, commit.Clone)
}

func TestCommit_Parse_ValidYAML(t *testing.T) {
	commit := NewCommit()

	format := "simplecontainer/v1/kind/containers/app/nginx"
	spec := `
image: nginx:latest
replicas: 3
ports:
  - container: "80/tcp"
    host: "8080:80/tcp"
`

	err := commit.Parse(format, spec)

	assert.NoError(t, err)
	assert.NotNil(t, commit.Format)
	assert.Equal(t, "nginx", commit.Format.GetName())
	assert.Equal(t, "app", commit.Format.GetGroup())
	assert.Equal(t, static.KIND_CONTAINERS, commit.Format.GetKind())
	assert.NotNil(t, commit.Patch)
}

func TestCommit_Parse_InvalidFormat(t *testing.T) {
	commit := NewCommit()

	format := "//invalid//"
	spec := `image: nginx:latest`

	err := commit.Parse(format, spec)

	assert.Error(t, err)
}

func TestCommit_Parse_InvalidYAML(t *testing.T) {
	commit := NewCommit()

	format := "smr.kind.containers.app.nginx"
	spec := `
invalid yaml:
  - this is: [not properly: formatted
`

	err := commit.Parse(format, spec)

	assert.Error(t, err)
}

func TestCommit_ToJson(t *testing.T) {
	commit := &Commit{
		Format:  f.New("smr", "kind", "containers", "app", "nginx"),
		Patch:   []byte(`{"image":"nginx:latest"}`),
		Message: "Update nginx image",
	}

	jsonBytes, err := commit.ToJson()

	assert.NoError(t, err)
	assert.NotNil(t, jsonBytes)

	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)
}

func TestCommit_FromJson(t *testing.T) {
	original := &Commit{
		Format:  f.New("simplecontainer", "v1", "kind", "containers", "app", "nginx"),
		Patch:   []byte(`{"image":"nginx:latest"}`),
		Message: "Update nginx",
	}

	jsonBytes, err := original.ToJson()
	assert.NoError(t, err)

	restored := NewCommit()
	err = restored.FromJson(jsonBytes)

	assert.NoError(t, err)
	assert.Equal(t, "nginx", restored.Format.GetName())
	assert.Equal(t, "app", restored.Format.GetGroup())
	assert.Equal(t, `{"image":"nginx:latest"}`, string(restored.Patch))
	assert.Equal(t, "Update nginx", restored.Message)
}

func TestCommit_FromJson_InvalidJSON(t *testing.T) {
	commit := NewCommit()

	err := commit.FromJson([]byte(`{invalid json`))

	assert.Error(t, err)
}

func TestCommit_GenerateClone_Containers(t *testing.T) {
	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_CONTAINERS, "app", "nginx"),
	}

	err := commit.GenerateClone()

	assert.NoError(t, err)
	assert.NotNil(t, commit.Clone)
	assert.NotNil(t, commit.Clone.Definition)
	assert.Equal(t, "nginx", commit.Clone.Definition.GetMeta().Name)
	assert.Equal(t, "app", commit.Clone.Definition.GetMeta().Group)
}

func TestCommit_GenerateClone_Configuration(t *testing.T) {
	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_CONFIGURATION, "app", "config"),
	}

	err := commit.GenerateClone()

	assert.NoError(t, err)
	assert.NotNil(t, commit.Clone)
	assert.Equal(t, static.KIND_CONFIGURATION, commit.Clone.Definition.GetKind())
}

func TestCommit_GenerateClone_Secret(t *testing.T) {
	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_SECRET, "app", "secret"),
	}

	err := commit.GenerateClone()

	assert.NoError(t, err)
	assert.NotNil(t, commit.Clone)
	assert.Equal(t, static.KIND_SECRET, commit.Clone.Definition.GetKind())
}

func TestCommit_GenerateClone_InvalidKind(t *testing.T) {
	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", "invalid-kind", "app", "test"),
	}

	err := commit.GenerateClone()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kind is not defined")
}

func TestCommit_ApplyPatch_SimpleField(t *testing.T) {
	// Create original definition
	original := definitions.New(static.KIND_CONTAINERS)
	containerDef := original.Definition.(*v1.ContainersDefinition)
	containerDef.Meta.Name = "nginx"
	containerDef.Meta.Group = "app"
	containerDef.Spec.Image = "nginx"
	containerDef.Spec.Tag = "1.0"
	containerDef.Spec.Replicas = 1

	// Create commit with patch
	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_CONTAINERS, "app", "nginx"),
		Patch:  []byte(`{"spec":{"tag":"2.0","replicas":3}}`),
	}

	err := commit.GenerateClone()
	assert.NoError(t, err)

	// Apply patch
	patchedBytes, err := commit.ApplyPatch(original.Definition)

	assert.NoError(t, err)
	assert.NotNil(t, patchedBytes)
	assert.NotEmpty(t, patchedBytes)

	// Verify patch was applied
	assert.Contains(t, string(patchedBytes), "tag: \"2.0\"")
	assert.Contains(t, string(patchedBytes), "replicas: 3")
}

func TestCommit_ApplyPatch_EmptyResult(t *testing.T) {
	original := definitions.New(static.KIND_CONTAINERS)
	containerDef := original.Definition.(*v1.ContainersDefinition)
	containerDef.Meta.Name = "nginx"
	containerDef.Meta.Group = "app"

	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_CONTAINERS, "app", "nginx"),
		Patch:  []byte(`{}`), // Empty patch
	}

	err := commit.GenerateClone()
	assert.NoError(t, err)

	_, err = commit.ApplyPatch(original.Definition)

	assert.Nil(t, err)
}

func TestCommit_ApplyPatch_InvalidPatch(t *testing.T) {
	original := definitions.New(static.KIND_CONTAINERS)
	containerDef := original.Definition.(*v1.ContainersDefinition)
	containerDef.Meta.Name = "nginx"
	containerDef.Meta.Group = "app"

	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_CONTAINERS, "app", "nginx"),
		Patch:  []byte(`{invalid json}`),
	}

	err := commit.GenerateClone()
	assert.NoError(t, err)

	_, err = commit.ApplyPatch(original.Definition)

	assert.Error(t, err)
}

func TestCommit_ApplyPatch_MessageGeneration(t *testing.T) {
	original := definitions.New(static.KIND_CONTAINERS)
	containerDef := original.Definition.(*v1.ContainersDefinition)
	containerDef.Meta.Name = "nginx"
	containerDef.Meta.Group = "app"
	containerDef.Spec.Image = "nginx"

	commit := &Commit{
		Format: f.New("simplecontainer", "v1", "kind", static.KIND_CONTAINERS, "app", "nginx"),
		Patch:  []byte(`{"spec":{"tag":"latest"}}`),
	}

	err := commit.GenerateClone()
	assert.NoError(t, err)

	_, err = commit.ApplyPatch(original.Definition)

	assert.NoError(t, err)
	assert.NotEmpty(t, commit.Message)
	assert.Contains(t, commit.Message, "applied patch")
	assert.Contains(t, commit.Message, "containers")
	assert.Contains(t, commit.Message, "app")
	assert.Contains(t, commit.Message, "nginx")
}

// ============================================================================
// FILE OPERATIONS TESTS
// ============================================================================

func TestCommit_WriteFile(t *testing.T) {
	commit := NewCommit()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.yaml")
	content := []byte("test: content\nkey: value\n")

	err := commit.WriteFile(filePath, content)

	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Verify content
	readContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Verify permissions (0600)
	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestCommit_WriteFile_InvalidPath(t *testing.T) {
	commit := NewCommit()

	// Try to write to non-existent directory
	filePath := "/nonexistent/directory/file.yaml"
	content := []byte("test")

	err := commit.WriteFile(filePath, content)

	assert.Error(t, err)
}

func TestCommit_WriteFile_Overwrite(t *testing.T) {
	commit := NewCommit()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.yaml")

	// Write first time
	content1 := []byte("original content")
	err := commit.WriteFile(filePath, content1)
	assert.NoError(t, err)

	// Overwrite
	content2 := []byte("new content")
	err = commit.WriteFile(filePath, content2)
	assert.NoError(t, err)

	// Verify new content
	readContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content2, readContent)
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestCommit_FullWorkflow(t *testing.T) {
	// 1. Parse YAML
	commit := NewCommit()
	format := "simplecontainer/v1/kind/containers/app/nginx"
	spec := `
spec:
  image: nginx
  tag: "1.0"
  replicas: 2
`

	err := commit.Parse(format, spec)
	assert.NoError(t, err)

	// 2. Generate clone
	err = commit.GenerateClone()
	assert.NoError(t, err)

	// 3. Create original definition
	original := definitions.New(static.KIND_CONTAINERS)
	containerDef := original.Definition.(*v1.ContainersDefinition)
	containerDef.Meta.Name = "nginx"
	containerDef.Meta.Group = "app"
	containerDef.Spec.Image = "nginx"
	containerDef.Spec.Tag = "0.9"
	containerDef.Spec.Replicas = 1

	// 4. Apply patch
	patchedBytes, err := commit.ApplyPatch(original.Definition)
	assert.NoError(t, err)
	assert.NotEmpty(t, patchedBytes)

	// 5. Write to file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nginx.yaml")
	err = commit.WriteFile(filePath, patchedBytes)
	assert.NoError(t, err)

	// 6. Verify file exists and has content
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "tag: \"1.0\"")
	assert.Contains(t, string(content), "replicas: 2")
}

func TestCommit_SerializationRoundTrip(t *testing.T) {
	original := &Commit{
		Format:  f.New("simplecontainer", "v1", "kind", "containers", "app", "nginx"),
		Patch:   []byte(`{"spec":{"image":"nginx:latest","replicas":5}}`),
		Message: "Update nginx configuration",
	}

	// Serialize
	jsonBytes, err := original.ToJson()
	assert.NoError(t, err)

	// Deserialize
	restored := NewCommit()
	err = restored.FromJson(jsonBytes)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, original.Format.GetName(), restored.Format.GetName())
	assert.Equal(t, original.Format.GetGroup(), restored.Format.GetGroup())
	assert.Equal(t, original.Format.GetKind(), restored.Format.GetKind())
	assert.Equal(t, original.Patch, restored.Patch)
	assert.Equal(t, original.Message, restored.Message)
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestCommit_Parse_EmptySpec(t *testing.T) {
	commit := NewCommit()

	format := "simplecontainer/v1/kind/containers/app/nginx"
	spec := ""

	err := commit.Parse(format, spec)

	assert.Error(t, err)
}

func TestCommit_Parse_ComplexNestedYAML(t *testing.T) {
	commit := NewCommit()

	format := "simplecontainer/v1/kind/containers/app/nginx"
	spec := `
image: nginx
spec:
  nested:
    deeply:
      nested:
        value: test
  array:
    - item1
    - item2
  map:
    key1: value1
    key2: value2
`

	err := commit.Parse(format, spec)

	assert.NoError(t, err)
	assert.NotNil(t, commit.Patch)
}

func TestCommit_WriteFile_EmptyContent(t *testing.T) {
	commit := NewCommit()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.yaml")

	err := commit.WriteFile(filePath, []byte{})

	assert.NoError(t, err)

	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}
