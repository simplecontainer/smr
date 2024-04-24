package manager

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"regexp"
	"smr/pkg/container"
	"smr/pkg/database"
	"smr/pkg/logger"
	"strings"
)

var Containers map[string]*container.Container = make(map[string]*container.Container, 3)
var Indexes map[string]string = make(map[string]string, 3)

func (mgr *Manager) Generate() {}

func (mgr *Manager) Load() {
	/*

		Indexes[static.GHOST] = mgr.getLatestName(static.GHOST)
		Indexes[static.NGINX] = mgr.getLatestName(static.NGINX)
		Indexes[static.MYSQL] = mgr.getLatestName(static.MYSQL)

	*/
}

func (mgr *Manager) Prepare(container *container.Container) bool {
	for keyOriginal, value := range container.Runtime.Configuration {
		// If {{ANYTHING_HERE}} is detected create template.Format type so that we can query the KV store if the format is valid
		format := database.FormatStructure{}

		logger.Log.Info("Trying to parse", zap.String("value", value.(string)))

		regexDetectBigBrackets := regexp.MustCompile(`{([^{{\n}}]*)}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value.(string), -1)

		if len(matches) > 0 {
			for index, _ := range matches {
				SplitByDot := strings.SplitN(matches[index][1], ".", 3)

				regexExtractGroupAndId := regexp.MustCompile(`([^\[\n\]]*)`)
				GroupAndIdExtractor := regexExtractGroupAndId.FindAllStringSubmatch(SplitByDot[1], -1)

				if len(GroupAndIdExtractor) > 1 {
					format := database.Format(SplitByDot[0], GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0], SplitByDot[2])

					if format.Identifier != "*" {
						format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), GroupAndIdExtractor[1][0])
					}

					key := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))
					val, err := database.Get(mgr.Badger, key)

					if err != nil {
						logger.Log.Error(val)
						return false
					}

					container.Runtime.Configuration[keyOriginal] = strings.Replace(container.Runtime.Configuration[keyOriginal].(string), fmt.Sprintf("{{%s}}", matches[index][1]), val, 1)
				}
			}
		} else {
			cleanKey := strings.TrimSpace(strings.Replace(keyOriginal, fmt.Sprintf("%s.%s.%s", "configuration", container.Static.Group, container.Static.GeneratedName), "", -1))
			format = database.Format("configuration", container.Static.Group, container.Static.GeneratedName, cleanKey)

			database.Put(mgr.Badger, fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key), value.(string))
		}
	}

	for keyOriginal, _ := range container.Runtime.Resources {
		// If {{ ANYTHING_HERE }} is detected create template.Format type so that we can query the KV store if the format is valid
		format := database.FormatStructure{}

		for k, dataEntry := range container.Runtime.Resources[keyOriginal].Data {
			regexDetectBigBrackets := regexp.MustCompile(`{([^{{\n}}]*)}`)
			matches := regexDetectBigBrackets.FindAllStringSubmatch(dataEntry, -1)

			logger.Log.Info("Trying to parse data in the resource", zap.String("value", k))

			if len(matches) > 0 {
				for index, _ := range matches {
					SplitByDot := strings.SplitN(matches[index][1], ".", 3)

					logger.Log.Info("Detected in the resource", zap.String("value", matches[index][1]))

					regexExtractGroupAndId := regexp.MustCompile(`([^\[\n\]]*)`)
					GroupAndIdExtractor := regexExtractGroupAndId.FindAllStringSubmatch(SplitByDot[1], -1)

					if len(GroupAndIdExtractor) > 1 {
						format = database.Format(SplitByDot[0], GroupAndIdExtractor[0][0], GroupAndIdExtractor[1][0], SplitByDot[2])

						if format.Identifier != "*" {
							format.Identifier = fmt.Sprintf("%s-%s", viper.GetString("project"), GroupAndIdExtractor[1][0])
						}

						key := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))
						val, err := database.Get(mgr.Badger, key)

						if err != nil {
							logger.Log.Error(val)
							return false
						}

						logger.Log.Info("Got value from the store", zap.String("value", key))

						container.Runtime.Resources[keyOriginal].Data[k] = strings.Replace(container.Runtime.Resources[keyOriginal].Data[k], fmt.Sprintf("{{%s}}", matches[index][1]), val, 1)
					}
				}
			}
		}
	}

	for index, value := range container.Static.Env {
		regexDetectBigBrackets := regexp.MustCompile(`{([^{{}}]*)}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

		if len(matches) > 0 {
			SplitByDot := strings.SplitN(matches[0][1], ".", 2)

			trimmedIndex := strings.TrimSpace(SplitByDot[1])

			if len(SplitByDot) > 1 && container.Runtime.Configuration[trimmedIndex] != nil {
				container.Static.Env[index] = strings.Replace(container.Static.Env[index], fmt.Sprintf("{{%s}}", matches[0][1]), container.Runtime.Configuration[trimmedIndex].(string), 1)
			}
		}
	}

	return true
}

func (mgr *Manager) Up() {
	/*
			if mgr.Config.Configuration.Mysql.Enabled {
				mgr.UpMysql()
			}

			if mgr.Config.Configuration.Ghost.Enabled {
				mgr.UpGhost()
			}

			if mgr.Config.Configuration.Nginx.Enabled {
				mgr.UpNginx()
			}
		}
		func (mgr *Manager) Down() {
			if mgr.Config.Configuration.Mysql.Enabled {
				mgr.DownMysql()
			}

			if mgr.Config.Configuration.Ghost.Enabled {
				mgr.DownGhost()
			}

			if mgr.Config.Configuration.Nginx.Enabled {
				mgr.DownNginx()
			}

	*/
}

func (mgr *Manager) UpgradeGhost() {
	/*
		mgr.InitUpgrade()

		viper.Set("scale-ghost", true)

		previousGhostIndex := mgr.getLatestName(static.GHOST)
		Indexes[static.GHOST] = mgr.GenerateName(static.GHOST)
		Indexes[static.NGINX] = mgr.GenerateName(static.NGINX)
		Indexes[static.MYSQL] = mgr.GenerateName(static.MYSQL)

		ghost := container.NewContainer(mgr.Runtime, Indexes[static.GHOST], mgr.Config.Configuration.Ghost.Image, mgr.Runtime.GhostUpgradeTag, mgr.GenerateNetworkName(), mgr.Config.Configuration.Ghost.FileMounts, mgr.Config.Configuration.Ghost.PortMappings, mgr.Config.Configuration.Ghost.ExposedPorts)
		mysql := container.NewContainer(mgr.Runtime, Indexes[static.MYSQL], mgr.Config.Configuration.Mysql.Image, mgr.Config.Configuration.Mysql.Tag, mgr.GenerateNetworkName(), mgr.Config.Configuration.Mysql.FileMounts, mgr.Config.Configuration.Mysql.PortMappings, mgr.Config.Configuration.Mysql.ExposedPorts)
		nginx := container.NewContainer(mgr.Runtime, Indexes[static.NGINX], mgr.Config.Configuration.Nginx.Image, mgr.Config.Configuration.Nginx.Tag, mgr.GenerateNetworkName(), mgr.Config.Configuration.Nginx.FileMounts, mgr.Config.Configuration.Nginx.PortMappings, mgr.Config.Configuration.Nginx.ExposedPorts)

		mysql.Get()

		Containers[Indexes[static.GHOST]] = ghost
		Containers[Indexes[static.NGINX]] = nginx

		oldDatabase := fmt.Sprintf("%s", strings.ReplaceAll(previousGhostIndex, "-", "_"))
		mgr.Config.Configuration.Mysql.Database = fmt.Sprintf("%s", strings.ReplaceAll(Indexes[static.GHOST], "-", "_"))
		mgr.Config.Configuration.Ghost.Host = Indexes[static.GHOST]
		mgr.Config.Configuration.Ghost.Tag = mgr.Runtime.GhostUpgradeTag

		mgr.Prepare(Containers[Indexes[static.NGINX]])

		mgr.SetupMysql(mysql.Runtime.IP, mgr.Config.Configuration.Mysql.Database)
		mgr.ExportDb(mysql, mgr.Runtime.PASSWORD, oldDatabase, fmt.Sprintf("%s/%s", mgr.Runtime.PROJECTDIR, static.UPGRADE_DIRTIME), static.MYSQL_DUMP)
		mgr.ImportDb(mysql, mgr.Config.Configuration.Mysql.User, mgr.Runtime.PASSWORD, mysql.Runtime.IP, mgr.Config.Configuration.Mysql.Database, fmt.Sprintf("%s/%s", mgr.Runtime.PROJECTDIR, static.UPGRADE_DIRTIME), static.MYSQL_DUMP)

		if utils.Confirm("Do you want to procede with upgrade?") {
			mgr.Prepare(Containers[Indexes[static.GHOST]])
			mgr.UpGhost()

			mgr.BlueGreen(nginx, container.Existing(previousGhostIndex))

			mgr.Config.Save(mgr.Runtime)

			if viper.GetBool("drop-old-db") {
				mgr.DropOldGhostDb(mgr.Config.Configuration.Mysql.User, mgr.Runtime.PASSWORD, mysql.Runtime.IP, oldDatabase)
			}
		} else {
			logger.Log.Info("Upgrade aborted! Configuration is not saved!")
		}

	*/
}

func (mgr *Manager) Watch() {
}

/*
func (mgr *Manager) GenerateName(name string) string {
	var project = mgr.Runtime.PROJECT

	if mgr.Config.Configuration.ProjectConfig.Parent != "" {
		if name != static.GHOST {
			project = mgr.Config.Configuration.ProjectConfig.Parent
		}
	}

	index := mgr.generateIndex(name)
	return fmt.Sprintf("%s-%s-%d", project, name, index)
}

func (mgr *Manager) getLatestName(name string) string {
	var project = mgr.Runtime.PROJECT

	if mgr.Config.Configuration.ProjectConfig.Parent != "" {
		if name != static.GHOST {
			project = mgr.Config.Configuration.ProjectConfig.Parent
		}
	}

	index := mgr.getPreviousIndex(name)
	return fmt.Sprintf("%s-%s-%d", project, name, index)
}

func (mgr *Manager) generateIndex(name string) int {
	var indexes []int = mgr.GetIndexes(name)
	var index int = 0

	if len(indexes) > 0 {
		sort.Ints(indexes)
		index = indexes[len(indexes)-1] + 1
	}

	switch name {
	case static.GHOST:
		if viper.Get("scale-ghost") == false {
			index = index - 1
		}
		break
	case static.MYSQL:
		if viper.Get("scale-mysql") == false {
			index = index - 1
		}
		break
	case static.NGINX:
		if viper.Get("scale-nginx") == false {
			index = index - 1
		}
		break
	}

	if index < 0 {
		index = 0
	}

	return index
}
func (mgr *Manager) getPreviousIndex(name string) int {
	var indexes []int = mgr.GetIndexes(name)
	var index int = 0

	if len(indexes) > 0 {
		sort.Ints(indexes)
		index = indexes[len(indexes)-1]
	}

	return index
}

func (mgr *Manager) GetIndexes(name string) []int {
	containers := container.GetContainers()

	var indexes = make([]int, 0)
	name = fmt.Sprintf("%s-%s", mgr.Runtime.PROJECT, name)

	if len(containers) > 0 {
		for _, container := range containers {
			for _, n := range container.Names {
				if strings.Contains(n, name) {
					fmt.Sprintf("%s contains %s", n, name)
					split := strings.Split(container.Names[0], "-")
					index, err := strconv.Atoi(split[len(split)-1])

					if err != nil {
						logger.Log.Fatal("Failed to convert string to int for index calculation")
					}

					indexes = append(indexes, index)
				}
			}
		}
	}

	if len(indexes) == 0 {
		switch name {
		case fmt.Sprintf("%s-%s", mgr.Runtime.PROJECT, static.GHOST):
			split := strings.Split(mgr.Config.Configuration.Ghost.Host, "-")
			index, err := strconv.Atoi(split[len(split)-1])

			if err != nil {
				logger.Log.Fatal("Failed to convert string to int for index calculation")
			}

			indexes = append(indexes, index)
			break
		}
	}

	return indexes
}
*/
