package static

const ROOTDIR string = "smr"
const CONFIGDIR string = "config"
const ROOTSMR = "smr"

const SMR string = "smr"

const DEFAULT_LOG_LEVEL = "info"

var STRUCTURE = []string{
	"config",
	"persistent",
	"persistent/smr",
	"persistent/etcd",
}

const SMR_SSH_HOME = "/home/smr-agent/.ssh/simplecontainer"

const SMR_ENDPOINT_NAME = "smr-agent"
const SMR_AGENT_URL = "smr-agent.cluster.private:1443"
const SMR_AGENT_DOMAIN = "smr-agent.cluster.private"
const SMR_LOCAL_DOMAIN string = "cluster.private"

const PLATFORM_DOCKER = "docker"

const CATEGORY_OBJECT = "badger"
const CATEGORY_ETCD = "etcd"
