package static

import (
	_ "embed"
)

const ROOTDIR string = "smr"
const CONFIGDIR string = "config"
const CONFIGFILE string = "smr.config"
const TEMPLATESDIR string = "templates"

const DOCKER_DNS_IP string = "127.0.0.11"
const SMR_LOCAL_DOMAIN string = "docker.private"

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
