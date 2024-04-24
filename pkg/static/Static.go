package static

import (
	_ "embed"
)

const ROOTDIR string = "smr"
const CONFIGDIR string = "config"
const CONFIGFILE string = "smr.config"
const TEMPLATESDIR string = "templates"

const SMR string = "smr"

const GHOST string = "ghost"
const NGINX string = "nginx"
const MYSQL string = "mysql"

const DOCKER_RUNNING string = "running"
const DOCKER_EXITED string = "exited"

var STRUCTURE = []string{
	"config",
	"persistent",
	"persistent/smr",
}

const SMR_ENDPOINT_NAME = "smr-agent"

//go:embed resources/git/version
var GHOSTMGR_VERSION string

//go:embed resources/nginx/development.conf
var NGINX_DEVELOPMENT string

//go:embed resources/nginx/production.conf
var NGINX_PRODUCTION string

//go:embed resources/ghost/development.conf
var GHOST_DEVELOPMENT string

//go:embed resources/ghost/production.conf
var GHOST_PRODUCTION string

//go:embed resources/shell/mysql_import.sh
var IMPORT_SCRIPT string
