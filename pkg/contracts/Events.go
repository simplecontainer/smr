package contracts

type Event interface {
	GetType() string
	GetTarget() string
	GetKind() string
	GetGroup() string
	GetName() string
	GetData() []byte
}

type PlatformEvent struct {
	NetworkID   string
	ContainerID string
	Target      string
	Group       string
	Name        string
	Managed     bool
	Type        string
}
