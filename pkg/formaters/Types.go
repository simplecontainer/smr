package formaters

type ContainerInformation struct {
	Group         string
	Name          string
	GeneratedName string
	Image         string
	ImageState    string
	Tag           string
	IPs           string
	Ports         string
	Dependencies  string
	DockerState   string
	SmrState      string
	NodeName      string
	NodeURL       string
	NodeID        uint64
	Recreated     bool
	LastUpdate    string
}
