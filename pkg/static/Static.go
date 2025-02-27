package static

const ROOTDIR string = "smr"
const CONFIGDIR string = "config"
const ROOTSMR = "smr"

const DEFAULT_LOG_LEVEL = "info"

var STRUCTURE = []string{
	"config",
	"persistent",
	"persistent/smr",
	"persistent/etcd",
}

const SMR_SSH_HOME = "/home/node/.ssh/simplecontainer"

const SMR_PREFIX = "simplecontainer.io/v1"
const SMR_ENDPOINT_NAME = "node"
const SMR_NODE_DOMAIN = "node.private"
const SMR_LOCAL_DOMAIN = "private"

const PLATFORM_DOCKER = "docker"
const PLATFORM_MOCKER = "mocker"

const CATEGORY_KIND = "kind"
const CATEGORY_STATE = "state"
const CATEGORY_ETCD = "etcd"
const CATEGORY_PLAIN = "plain"
const CATEGORY_EVENT = "event"
const CATEGORY_SECRET = "secret"
const CATEGORY_DNS = "dns"
const CATEGORY_INVALID = "invalid"

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
const KIND_SECRET = "secret"
const KIND_CUSTOM = "custom"

const REMOVE_KIND = "remove"

const RESPONSE_SCHEDULED = "action accepted and scheduled for action"
const RESPONSE_APPLIED = "object is applied"
const RESPONSE_BAD_REQUEST = "request sent is invalid"
const RESPONSE_DELETED = "object is deleted"
const RESPONSE_RESTART = "object is restarted"
const RESPONSE_REFRESHED = "object is refreshed"
const RESPONSE_SYNCED = "object is synced"
const RESPONSE_NOT_FOUND = "object is not found"
const RESPONSE_INTERNAL_ERROR = "object action errored on the server"
