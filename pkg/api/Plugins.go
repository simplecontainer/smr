package api

import (
	"errors"
	"fmt"
	"github.com/qdnqn/smr/pkg/implementations"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func (api *Api) StartPlugins() {
	plugins := make([]string, 0)

	implementationsDir := fmt.Sprintf("%s/%s", api.Config.Configuration.Environment.Root, "implementations")

	files, _ := os.ReadDir(implementationsDir)
	path, _ := filepath.Abs(implementationsDir)

	for _, file := range files {
		plugins = append(plugins, filepath.Join(path, file.Name()))
	}

	for _, pluginPath := range plugins {
		pluginName := filepath.Base(pluginPath)
		pluginName = strings.TrimSuffix(pluginName, ".so")

		plugin, err := getPluginInstance(api.Config.Configuration.Environment.Root, "implementations", pluginName)

		if err != nil {
			panic(err)
		}

		if plugin != nil {
			ImplementationInternal, err := plugin.Lookup(cases.Title(language.English).String(pluginName))

			if err != nil {
				panic(errors.New("plugin lookup failed"))
			} else {
				pl, ok := ImplementationInternal.(implementations.Implementation)

				if !ok {
					panic(errors.New("casting plugin to implementation failed"))
				} else {
					err = pl.Start(api.Manager)

					if err != nil {
						panic(err)
					}
				}
			}
		} else {
			panic(errors.New("plugin is nil"))
		}
	}
}

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
