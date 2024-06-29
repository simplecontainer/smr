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

var CLIENT_CONTEXT_DIR = "contexts"

var CLIENT_STRUCTURE = []string{
	CLIENT_CONTEXT_DIR,
}

const SMR_ENDPOINT_NAME = "smr-agent"

//go:embed resources/git/version
var SMR_VERSION string

// Container statuses

const STATUS_BACKOFF int = 0
const STATUS_READY int = 1
const STATUS_DEPENDS_SOLVED int = 2
const STATUS_HEALTHY int = 3
const STATUS_RUNNING int = 4
const STATUS_RECONCILING int = 5
const STATUS_DRIFTED int = 6
const STATUS_PENDING_DELETE int = 7
const STATUS_CREATED int = 8
const STATUS_READINESS int = 9
const STATUS_READINESS_FAILED int = 10
