package implementation

import "fmt"

func (gitops *Gitops) GetGroup() string {
	return gitops.Definition.Meta.Group
}

func (gitops *Gitops) GetName() string {
	return gitops.Definition.Meta.Name
}

func (gitops *Gitops) GetGroupIdentifier() string {
	return fmt.Sprintf("%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
}
