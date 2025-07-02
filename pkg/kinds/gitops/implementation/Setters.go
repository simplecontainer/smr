package implementation

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/pkg/packer"
)

func (gitops *Gitops) SetCommit(commit *object.Commit, err error) error {
	if err != nil {
		return err
	}

	gitops.Gitops.Commit = commit
	return nil
}

func (gitops *Gitops) SetPack(pack *packer.Pack, err error) error {
	if err != nil {
		return err
	}

	gitops.Gitops.Pack = pack
	return nil
}

func (gitops *Gitops) SetForceSync(value bool) { gitops.Gitops.ForceSync = value }

func (gitops *Gitops) SetForceClone(value bool) { gitops.Gitops.ForceClone = value }
