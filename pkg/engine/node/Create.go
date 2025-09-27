package node

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"net"
)

func Create(api iapi.Api) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())

	_, err := bootstrap.CreateProject(static.ROOTSMR, environment, 0750)

	if err != nil {
		panic(err)
	}

	api.GetConfig().Platform = viper.GetString("platform")
	api.GetConfig().HostPort.Host, api.GetConfig().HostPort.Port, err = net.SplitHostPort(viper.GetString("listen"))

	if err != nil {
		panic(err)
	}

	api.GetConfig().NodeName = viper.GetString("node")

	api.GetConfig().NodeImage = viper.GetString("image")
	api.GetConfig().NodeTag = viper.GetString("tag")
	api.GetConfig().Certificates.Domains = configuration.NewDomains([]string{viper.GetString("domain")})
	api.GetConfig().Certificates.IPs = configuration.NewIPs([]string{viper.GetString("ip")})

	// Internal domains needed
	api.GetConfig().Certificates.Domains.Add("localhost")
	api.GetConfig().Certificates.Domains.Add(fmt.Sprintf("%s.%s", static.SMR_ENDPOINT_NAME, static.SMR_LOCAL_DOMAIN))

	// Internal IPs needed
	api.GetConfig().Certificates.IPs.Add("127.0.0.1")

	api.GetConfig().KVStore = &configuration.KVStore{
		Cluster: nil,
		Node:    nil,
		URL:     viper.GetString("url"),
		Join:    viper.GetBool("join"),
		Peer:    viper.GetString("peer"),
	}

	api.GetConfig().Ports = &configuration.Ports{
		Control: viper.GetString("port.control"),
		Overlay: viper.GetString("port.overlay"),
		Etcd:    viper.GetString("port.etcd"),
	}

	err = startup.Save(api.GetConfig(), environment, 0750)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println(fmt.Sprintf("config created and saved at %s/config/config.yaml", environment.NodeDirectory))
}
