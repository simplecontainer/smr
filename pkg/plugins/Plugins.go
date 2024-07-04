package plugins

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/implementations"
	"github.com/simplecontainer/smr/pkg/manager"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func StartPlugins(implementationsRootDir string, mgr *manager.Manager) {
	plugins := make([]string, 0)
	plugins = append(plugins, "hub.so")

	implementationsDir := fmt.Sprintf("%s/%s", implementationsRootDir, "implementations")

	files, _ := os.ReadDir(implementationsDir)

	for _, file := range files {
		if file.Name() != plugins[0] {
			plugins = append(plugins, file.Name())
		}
	}

	for _, pluginName := range plugins {
		pl := GetPlugin(implementationsRootDir, pluginName)

		if pl != nil {
			pl.Start(mgr)
		}
	}
}

func GetPlugin(implementationsRootDir string, pluginWanted string) implementations.Implementation {
	plugins := make([]string, 0)

	implementationsDir := fmt.Sprintf("%s/%s", implementationsRootDir, "implementations")

	files, _ := os.ReadDir(implementationsDir)
	path, _ := filepath.Abs(implementationsDir)

	for _, file := range files {
		if file.Name() == pluginWanted {
			plugins = append(plugins, filepath.Join(path, file.Name()))
		}
	}

	for _, pluginPath := range plugins {
		pluginName := filepath.Base(pluginPath)
		pluginName = strings.TrimSuffix(pluginName, ".so")

		plugin, err := GetPluginInstance(implementationsRootDir, "implementations", pluginName)

		if err != nil {
			panic(err)
		}

		if plugin != nil {
			ImplementationInternal, err := plugin.Lookup(cases.Title(language.English).String(pluginName))

			if err != nil {
				panic(err)
			} else {
				pl, ok := ImplementationInternal.(implementations.Implementation)

				if !ok {
					panic(errors.New("casting plugin to implementation failed"))
				} else {
					return pl
				}
			}
		} else {
			panic(errors.New("plugin is nil"))
		}
	}

	return nil
}

func GetPluginShared(implementationsRootDir string, pluginWanted string) implementations.ImplementationShared {
	plugins := make([]string, 0)

	implementationsDir := fmt.Sprintf("%s/%s", implementationsRootDir, "implementations")

	files, _ := os.ReadDir(implementationsDir)
	path, _ := filepath.Abs(implementationsDir)

	for _, file := range files {
		if file.Name() == pluginWanted {
			plugins = append(plugins, filepath.Join(path, file.Name()))
		}
	}

	for _, pluginPath := range plugins {
		pluginName := filepath.Base(pluginPath)
		pluginName = strings.TrimSuffix(pluginName, ".so")

		plugin, err := GetPluginInstance(implementationsRootDir, "implementations", pluginName)

		if err != nil {
			panic(err)
		}

		if plugin != nil {
			ImplementationInternal, err := plugin.Lookup(cases.Title(language.English).String(pluginName))

			if err != nil {
				panic(err)
			} else {
				pl, ok := ImplementationInternal.(implementations.ImplementationShared)

				if !ok {
					panic(errors.New("casting plugin to implementation failed"))
				} else {
					return pl
				}
			}
		} else {
			panic(errors.New("plugin is nil"))
		}
	}

	return nil
}

func GetPluginInstance(projectDir string, typ string, name string) (*plugin.Plugin, error) {
	var pluginInstance *plugin.Plugin
	var err error

	pluginInstance, err = plugin.Open(fmt.Sprintf("%s/%s/%s.so", projectDir, typ, name))

	return pluginInstance, err
}
