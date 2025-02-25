package platform

type Event struct {
	NetworkID   string
	ContainerID string
	Kind        string
	Target      string
	Group       string
	Name        string
	Managed     bool
	Type        string
	Data        []byte
}
