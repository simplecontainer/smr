package configuration

type Configuration struct {
	Platform    string       `yaml:"platform"`
	Target      string       `default:"development" yaml:"target"`
	Root        string       `yaml:"root"`
	OptRoot     string       `yaml:"optroot"`
	Domain      string       `yaml:"domain"`
	ExternalIP  string       `yaml:"externalIP"`
	CommonName  string       `yaml:"commonName"`
	HostHome    string       `yaml:"hostHome"`
	Environment *Environment `yaml:"-"`
	Flags       Flags        `yaml:"-"`
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
