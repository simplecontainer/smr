package configuration

type Configuration struct {
	Platform       string       `yaml:"platform"`
	OverlayNetwork string       `yaml:overlaynetwork`
	Agent          string       `yaml:"agent"`
	Port           int          `yaml:"port"`
	KVStore        *KVStore     `yaml:"kvstore"`
	Target         string       `default:"development" yaml:"target"`
	Root           string       `yaml:"root"`
	OptRoot        string       `yaml:"optroot"`
	Domain         string       `yaml:"domain"`
	ExternalIP     string       `yaml:"externalIP"`
	CommonName     string       `yaml:"CN"`
	HostHome       string       `yaml:"home"`
	Node           string       `yaml:"node"`
	Environment    *Environment `yaml:"-"`
	Flags          Flags        `yaml:"-"`
}

type KVStore struct {
	Cluster     []string `yaml:"cluster"`
	Node        uint64   `yaml:"node"`
	URL         string   `yaml:"url"`
	JoinCluster bool     `yaml:"join"`
}

type Flags struct {
	Opt     bool
	Verbose bool
}

type Environment struct {
	HOMEDIR    string
	OPTDIR     string
	PROJECTDIR string
	PROJECT    string
	AGENTIP    string
}
