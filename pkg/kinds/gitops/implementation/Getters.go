package implementation

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
)

func (gitops *Gitops) GetGroup() string {
	return gitops.Definition.Meta.Group
}

func (gitops *Gitops) GetName() string {
	return gitops.Definition.Meta.Name
}

func (gitops *Gitops) GetStatus() *status.Status {
	return gitops.Status
}

func (gitops *Gitops) GetGroupIdentifier() string {
	return fmt.Sprintf("%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
}
