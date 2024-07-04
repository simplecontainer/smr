package configuration

type Configuration struct {
	Target      string `default:"development" json:"target"`
	Root        string `json:"root"`
	Environment *Environment
	Flags       Flags
}

type Flags struct {
	Daemon        bool
	DaemonSecured bool
	DaemonDomain  string
	Verbose       bool
	OptMode       bool
}

type Environment struct {
	HOMEDIR    string
	OPTDIR     string
	PROJECTDIR string
	PROJECT    string
	AGENTIP    string
}
