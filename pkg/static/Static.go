package static

// Directory Constants
const (
	ROOTDIR      = "smr"
	CONFIGDIR    = "config"
	ROOTSMR      = "smr"
	SMR_SSH_HOME = "/home/node/.ssh/simplecontainer"
)

// Default Log Level
const DEFAULT_LOG_LEVEL = "info"

// Structure Paths
var STRUCTURE = []string{
	"config",
	"persistent",
	"persistent/smr",
	"persistent/etcd",
}

// SMR Config Constants
const (
	SMR_PREFIX        = "simplecontainer.io/v1"
	SMR_ENDPOINT_NAME = "node"
	SMR_NODE_DOMAIN   = "node.private"
	SMR_LOCAL_DOMAIN  = "private"
)

// Cluster Constants
const CLUSTER_NETWORK = "cluster"

// Platform Constants
const (
	PLATFORM_DOCKER = "docker"
	PLATFORM_MOCKER = "mocker"
)

// Category Constants
const (
	CATEGORY_KIND    = "kind"
	CATEGORY_STATE   = "state"
	CATEGORY_ETCD    = "etcd"
	CATEGORY_PLAIN   = "plain"
	CATEGORY_EVENT   = "event"
	CATEGORY_SECRET  = "secret"
	CATEGORY_DNS     = "dns"
	CATEGORY_INVALID = "invalid"
)

// Signal Constants
const (
	SIGTERM = "SIGTERM"
	SIGKILL = "SIGKILL"
)

// Kind Constants
const (
	KIND_CONTAINER     = "container"
	KIND_CONTAINERS    = "containers"
	KIND_CONFIGURATION = "configuration"
	KIND_RESOURCE      = "resource"
	KIND_CERTKEY       = "certkey"
	KIND_HTTPAUTH      = "httpauth"
	KIND_GITOPS        = "gitops"
	KIND_NETWORK       = "network"
	KIND_SECRET        = "secret"
	KIND_CUSTOM        = "custom"
)

// State Constants
const STATE_KIND = "state"
const REMOVE_KIND = "remove"

// Response Constants
const (
	RESPONSE_SCHEDULED      = "action accepted and scheduled for action"
	RESPONSE_APPLIED        = "object is applied"
	RESPONSE_BAD_REQUEST    = "request sent is invalid"
	RESPONSE_DELETED        = "object is deleted"
	RESPONSE_RESTART        = "object is restarted"
	RESPONSE_REFRESHED      = "object is refreshed"
	RESPONSE_SYNCED         = "object is synced"
	RESPONSE_NOT_FOUND      = "object is not found"
	RESPONSE_INTERNAL_ERROR = "object action errored on the server"
)
