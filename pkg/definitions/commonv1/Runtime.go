package commonv1

type Runtime struct {
	Owner Owner  `json:"owner"`
	Node  string `json:"node"`
}

func (runtime *Runtime) SetNode(node string) {
	runtime.Node = node
}

func (runtime *Runtime) GetNode() string {
	return runtime.Node
}

func (runtime *Runtime) SetOwner(kind string, group string, name string) {
	runtime.Owner.Kind = kind
	runtime.Owner.Group = group
	runtime.Owner.Name = name
}

func (runtime *Runtime) GetOwner() Owner {
	return runtime.Owner
}
