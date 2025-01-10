package commonv1

type Owner struct {
	Kind  string
	Group string
	Name  string
}

func (owner Owner) IsEmpty() bool {
	return owner.Group != "" && owner.Name != ""
}
