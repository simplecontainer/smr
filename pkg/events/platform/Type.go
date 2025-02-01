package platform

type Event struct {
	NetworkID   string
	ContainerID string
	Target      string
	Group       string
	Name        string
	Managed     bool
	Type        string
}
