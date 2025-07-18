package commonv1

type Owner struct {
	Kind  string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Group string `json:"group,omitempty" yaml:"group,omitempty"`
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (o *Owner) GetKind() string {
	return o.Kind
}

func (o *Owner) SetKind(kind string) {
	o.Kind = kind
}

func (o *Owner) GetGroup() string {
	return o.Group
}

func (o *Owner) SetGroup(group string) {
	o.Group = group
}

func (o *Owner) GetName() string {
	return o.Name
}

func (o *Owner) SetName(name string) {
	o.Name = name
}

func (o *Owner) IsEmpty() bool {
	return o.Group == "" && o.Name == ""
}

func (o *Owner) IsEqual(cmp *Owner) bool {
	return o.Kind == cmp.Kind && o.Name == cmp.Name && o.Group == cmp.Group
}
