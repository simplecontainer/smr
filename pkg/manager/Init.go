package manager

import (
	_ "github.com/go-sql-driver/mysql"
	"os"
	"smr/pkg/logger"
)

func (mgr *Manager) Init() {
	/*
		project := mgr.Runtime.PROJECTDIR

		for _, path := range static.STRUCTURE {
			dir := fmt.Sprintf("%s/%s", project, path)

			if _, err := os.Stat(dir); os.IsNotExist(err) {
				logger.Log.Info("Creating directory.", zap.String("Directory", dir))

				err := os.MkdirAll(dir, 0750)
				if err != nil {
					logger.Log.Fatal(err.Error())

					err := os.RemoveAll(project)
					if err != nil {
						logger.Log.Fatal(err.Error())
					}
				}
			}
		}

		developmentConfig := getDevelopmentConfig(mgr.Runtime.PROJECTDIR, mgr.Runtime.PROJECT)

		dump, err := yaml.Marshal(developmentConfig)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s", project, static.CONFIGDIR, static.DEVELOPMENTCONFIG), dump, 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		productionConfig := getProductionConfig(mgr.Runtime.PROJECTDIR, mgr.Runtime.PROJECT)

		dump, err = yaml.Marshal(productionConfig)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s", project, static.CONFIGDIR, static.PRODUCTIONCONFIG), dump, 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s/development.conf", project, static.TEMPLATESDIR, static.GHOST), []byte(static.GHOST_DEVELOPMENT), 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s/production.conf", project, static.TEMPLATESDIR, static.GHOST), []byte(static.GHOST_PRODUCTION), 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s/development.conf", project, static.TEMPLATESDIR, static.NGINX), []byte(static.NGINX_DEVELOPMENT), 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s/production.conf", project, static.TEMPLATESDIR, static.NGINX), []byte(static.NGINX_PRODUCTION), 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		if mgr.Runtime.PASSWORD == "" {
			mgr.Runtime.PASSWORD = utils.RandString(32)
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s/.password", project, static.TEMPLATESDIR, static.MYSQL), []byte(fmt.Sprintf("%s", mgr.Runtime.PASSWORD)), 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		} else {
			logger.Log.Info(fmt.Sprintf("======================================================================================="))
			logger.Log.Info(fmt.Sprintf("Mysql generated credentials. Username: %s, Password %s", "root", mgr.Runtime.PASSWORD))
			logger.Log.Info(fmt.Sprintf("It will be deleted at the end of the program. Be sure to save it!"))
			logger.Log.Info(fmt.Sprintf("======================================================================================="))
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s/%s/.envs", project, static.TEMPLATESDIR, static.MYSQL), []byte(fmt.Sprintf("%s=%s", "MYSQL_ROOT_PASSWORD", "{{ .Runtime.PASSWORD }}")), 0750)
		if err != nil {
			logger.Log.Fatal(err.Error())
		} else {
			logger.Log.Info(fmt.Sprintf("Added MYSQL_ROOT_PASSWORD to the .envs"))
		}

	*/
}

func (mgr *Manager) Destroy() {
	project := mgr.Runtime.PROJECTDIR

	err := os.RemoveAll(project)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
