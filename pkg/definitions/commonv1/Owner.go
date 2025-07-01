package commonv1

type Owner struct {
	Kind  string `json:"kind,omitempty"`
	Group string `json:"group,omitempty"`
	Name  string `json:"name,omitempty"`
}

func (owner Owner) IsEmpty() bool {
	return owner.Group == "" && owner.Name == ""
}

func (owner Owner) IsEqual(o Owner) bool {
	return owner.Kind == o.Kind && owner.Name == o.Name && owner.Group == o.Group
}
