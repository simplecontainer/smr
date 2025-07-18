package implementation

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation/internal"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/packer"
)

func (gitops *Gitops) GetDefinition() idefinitions.IDefinition {
	return gitops.Gitops.definition
}

func (gitops *Gitops) GetGroup() string {
	return gitops.Gitops.definition.Meta.Group
}

func (gitops *Gitops) GetAutoSync() bool { return gitops.Gitops.AutomaticSync }

func (gitops *Gitops) GetDirectory() string { return gitops.Gitops.DirectoryPath }

func (gitops *Gitops) GetForceSync() bool { return gitops.Gitops.ForceSync }

func (gitops *Gitops) GetForceClone() bool { return gitops.Gitops.ForceClone }

func (gitops *Gitops) GetName() string {
	return gitops.Gitops.definition.Meta.Name
}

func (gitops *Gitops) GetStatus() *status.Status {
	return gitops.Gitops.Status
}

func (gitops *Gitops) GetPack() *packer.Pack { return gitops.Gitops.Pack }

func (gitops *Gitops) GetGit() *internal.Git { return gitops.Gitops.Git }

func (gitops *Gitops) GetCommit() *object.Commit { return gitops.Gitops.Commit }

func (gitops *Gitops) GetGroupIdentifier() string {
	return fmt.Sprintf("%s/%s", gitops.Gitops.definition.Meta.Group, gitops.Gitops.definition.Meta.Name)
}

func (gitops *Gitops) GetQueue() *QueueTS {
	return gitops.Gitops.PatchQueue
}
