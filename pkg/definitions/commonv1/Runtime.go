package commonv1

type Runtime struct {
	Owner    Owner  `json:"owner,omitempty"`
	Node     uint64 `json:"node,omitempty"`
	NodeName string `json:"nodeName,omitempty"`
}

func (runtime *Runtime) SetNode(node uint64) {
	runtime.Node = node
}

func (runtime *Runtime) GetNode() uint64 {
	return runtime.Node
}

func (runtime *Runtime) SetNodeName(nodeName string) {
	runtime.NodeName = nodeName
}

func (runtime *Runtime) GetNodeName() string {
	return runtime.NodeName
}

func (runtime *Runtime) SetOwner(kind string, group string, name string) {
	runtime.Owner.Kind = kind
	runtime.Owner.Group = group
	runtime.Owner.Name = name
}

func (runtime *Runtime) GetOwner() Owner {
	return runtime.Owner
}
