package runtime

import "net"

type Runtime struct {
	HOMEDIR         string
	OPTDIR          string
	PROJECTDIR      string
	PROJECT         string
	PASSWORD        string
	AGENTIP         net.IP
	GhostUpgradeTag string
}
