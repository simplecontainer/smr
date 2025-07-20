package implementation

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/packer"
	"os"
	"path/filepath"
	"strings"
)

func (gitops *Gitops) GetDefinition() idefinitions.IDefinition {
	return gitops.Gitops.definition
}

func (gitops *Gitops) GetGroup() string {
	return gitops.Gitops.Group
}

func (gitops *Gitops) GetAutoSync() bool { return gitops.Gitops.AutomaticSync }

func (gitops *Gitops) GetDirectory() string { return gitops.Gitops.DirectoryPath }

func (gitops *Gitops) GetForceSync() bool { return gitops.Gitops.ForceSync }

func (gitops *Gitops) GetForceClone() bool { return gitops.Gitops.ForceClone }

func (gitops *Gitops) GetName() string {
	return gitops.Gitops.Name
}

func (gitops *Gitops) GetStatus() *status.Status {
	return gitops.Gitops.Status
}

func (gitops *Gitops) GetPack() *packer.Pack { return gitops.Gitops.Pack }

func (gitops *Gitops) GetGit() *internal.Git { return gitops.Gitops.Git }

func (gitops *Gitops) GetCommit() *object.Commit { return gitops.Gitops.Commit }

func (gitops *Gitops) GetFilePath(file string) (*FilePath, error) {
	gitDir := filepath.Clean(gitops.Gitops.Git.Directory)
	dirPath := strings.Trim(filepath.Clean(gitops.Gitops.DirectoryPath), "/")
	fileName := strings.Trim(filepath.Clean(file), "/")

	if strings.Contains(fileName, "..") {
		return nil, fmt.Errorf("invalid file name: path traversal not allowed")
	}

	var relativePath string
	if dirPath == "" || dirPath == "." {
		relativePath = filepath.Join("definitions", fileName)
	} else {
		relativePath = filepath.Join(dirPath, "definitions", fileName)
	}

	absolutePath := filepath.Join(gitDir, relativePath)

	if absGitDir, err := filepath.Abs(gitDir); err == nil {
		if absPath, err := filepath.Abs(absolutePath); err == nil {
			if !strings.HasPrefix(absPath, absGitDir) {
				return nil, fmt.Errorf("file path is outside git directory")
			}
		}
	}

	if _, err := os.Stat(absolutePath); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return &FilePath{
		Absolute: absolutePath,
		Relative: relativePath,
	}, nil
}

func (gitops *Gitops) GetGroupIdentifier() string {
	return fmt.Sprintf("%s/%s", gitops.Gitops.definition.Meta.Group, gitops.Gitops.definition.Meta.Name)
}

func (gitops *Gitops) GetQueue() *QueueTS {
	return gitops.Gitops.PatchQueue
}
