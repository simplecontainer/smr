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
const SMR_AGENT_URL = "smr-agent.private:1443"
const SMR_AGENT_DOMAIN = "smr-agent.private"
const SMR_LOCAL_DOMAIN string = "private"

const PLATFORM_DOCKER = "docker"

const CATEGORY_OBJECT = 0x0
const CATEGORY_ETCD = 0x1
const CATEGORY_PLAIN = 0x2
const CATEGORY_EVENT = 0x3
const CATEGORY_SECRET = 0x4
const CATEGORY_OBJECT_DELETE = 0x5
const CATEGORY_INVALID = 0x9

const CATEGORY_OBJECT_STRING = "object"
const CATEGORY_ETCD_STRING = "etcd"
const CATEGORY_PLAIN_STRING = "plain"
const CATEGORY_EVENT_STRING = "event"
const CATEGORY_SECRET_STRING = "secret"
const CATEGORY_INVALID_STRING = "invalid"

const SIGTERM = "SIGTERM"
const SIGKILL = "SIGKILL"

const KIND_CONTAINER = "container"
const KIND_CONTAINERS = "containers"
const KIND_CONFIGURATION = "configuration"
const KIND_RESOURCE = "resource"
const KIND_CERTKEY = "certkey"
const KIND_HTTPAUTH = "httpauth"
const KIND_GITOPS = "gitops"
const KIND_NETWORK = "network"

const STATUS_RESPONSE_APPLIED = "object is applied"
const STATUS_RESPONSE_BAD_REQUEST = "request sent is invalid"
const STATUS_RESPONSE_DELETED = "object is deleted"
const STATUS_RESPONSE_RESTART = "object is restarted"
const STATUS_RESPONSE_REFRESHED = "object is refreshed"
const STATUS_RESPONSE_SYNCED = "object is synced"
const STATUS_RESPONSE_NOT_FOUND = "object is not found"
const STATUS_RESPONSE_INTERNAL_ERROR = "object action errored on the server"
