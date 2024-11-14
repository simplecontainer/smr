package startup

import (
	"flag"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"os"
)

func Load(in io.Reader) (*configuration.Configuration, error) {
	configObj := configuration.NewConfig()

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(in)

	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(configObj)

	if err != nil {
		return nil, err
	}

	configObj.Environment = GetEnvironmentInfo()

	return configObj, err
}

func Save(configObj *configuration.Configuration, out io.Writer) error {
	yamlObj, err := yaml.Marshal(*configObj)

	if err != nil {
		panic(err)
	}

	_, err = out.Write(yamlObj)

	if err != nil {
		return err
	}

	return nil
}

func SetFlags() {
	flag.String("project", "", "Project name")
	flag.Bool("opt", false, "Run in opt mode - do it only in containers")
	flag.Int("port", 1443, "SMR TLS port")
	flag.Bool("verbose", false, "Verbose output")
	flag.String("agent", "", "Agent container name")
	flag.String("cluster", "", "SMR Cluster")
	flag.String("overlay", "10.10.0.0/16", "Overlay network for flannel to use")
	flag.Int("node", 0, "Distributed KVStore Node ID")
	flag.Bool("join", false, "Join the cluster")
	flag.Bool("restore", false, "Restore cluster")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

func ReadFlags(configObj *configuration.Configuration) {
	configObj.Flags.Opt = viper.GetBool("opt")
	configObj.Flags.Verbose = viper.GetBool("verbose")
}

func GetEnvironmentInfo() *configuration.Environment {
	HOMEDIR, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}

	OPTDIR := "/opt/smr"

	if viper.GetBool("opt") {
		if _, err = os.Stat(OPTDIR); err != nil {
			if err = os.Mkdir(OPTDIR, os.FileMode(0750)); err != nil {
				panic(err.Error())
			}
		}
	}

	if err != nil {
		panic(err)
	}

	return &configuration.Environment{
		HOMEDIR:    HOMEDIR,
		OPTDIR:     OPTDIR,
		PROJECT:    static.PROJECT,
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
