package static

import (
	_ "embed"
)

const ROOTDIR string = "smr"
const CONFIGDIR string = "config"
const PROJECT = "smr"
const SMR_LOCAL_DOMAIN string = "docker.private"

const SMR string = "smr"

const DEFAULT_LOG_LEVEL = "info"

var STRUCTURE = []string{
	"config",
	"persistent",
	"persistent/smr",
}

const SMR_ENDPOINT_NAME = "smr-agent"

//go:embed resources/git/version
var SMR_VERSION string
