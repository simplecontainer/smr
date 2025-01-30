package configuration

import (
	"fmt"
	ips "github.com/simplecontainer/smr/pkg/network/ip"
	"github.com/simplecontainer/smr/pkg/static"
)

func NewConfig() *Configuration {
	IPs, err := ips.NewfromEtcHosts()

	if err != nil {
		panic(err)
	}

	return &Configuration{
		Environment: &Environment{
			Home:          "/home/node",
			NodeDirectory: fmt.Sprintf("%s/%s", "/home/node", static.ROOTDIR),
			NodeIP:        IPs.IPs[len(IPs.IPs)-1].String(),
		},
		Certificates: &Certificates{},
	}
}
