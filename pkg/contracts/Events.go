package contracts

type PlatformEvent struct {
	NetworkID   string
	ContainerID string
	Group       string
	Name        string
	Managed     bool
	Type        string
}
