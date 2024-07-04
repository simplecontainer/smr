package startup

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"os"
)

func GetEnvironmentInfo() *configuration.Environment {
	HOMEDIR, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}

	OPTDIR := "/opt/smr"

	if _, err := os.Stat(OPTDIR); err != nil {
		if err = os.Mkdir(OPTDIR, os.FileMode(0750)); err != nil {
			panic(err.Error())
		}
	}

	return &configuration.Environment{
		HOMEDIR:    HOMEDIR,
		OPTDIR:     OPTDIR,
		PROJECT:    fmt.Sprintf("%s", static.PROJECT),
		PROJECTDIR: fmt.Sprintf("%s/%s/%s", HOMEDIR, static.ROOTDIR, static.PROJECT),
		AGENTIP:    GetOutboundIP().String(),
	}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
