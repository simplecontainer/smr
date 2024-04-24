package api

import (
	"fmt"
	"github.com/spf13/viper"
	"plugin"
)

func getPluginInstance(projectDir string, typ string, name string) (*plugin.Plugin, error) {
	var pluginInstance *plugin.Plugin
	var err error

	if viper.GetBool("optmode") {
		pluginInstance, err = plugin.Open(fmt.Sprintf("%s/%s/%s/%s.so", projectDir, typ, name, name))
	} else {
		pluginInstance, err = plugin.Open(fmt.Sprintf("%s/%s/%s.so", typ, name, name))
	}

	return pluginInstance, err
}
