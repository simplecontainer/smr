package api

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/ikinds"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/relations"
	"github.com/simplecontainer/smr/pkg/version"
	"github.com/simplecontainer/smr/pkg/wss"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/embed"
)

func (a *Api) GetServer() *embed.Etcd  { return a.Server }
func (a *Api) SetServer(e *embed.Etcd) { a.Server = e }

func (a *Api) GetEtcd() *clientv3.Client  { return a.Etcd }
func (a *Api) SetEtcd(c *clientv3.Client) { a.Etcd = c }

func (a *Api) GetLeaseIdentifier() *clientv3.LeaseGrantResponse  { return a.LeaseIdentifier }
func (a *Api) SetLeaseIdentifier(l *clientv3.LeaseGrantResponse) { a.LeaseIdentifier = l }

func (a *Api) GetUser() *authentication.User  { return a.User }
func (a *Api) SetUser(u *authentication.User) { a.User = u }

func (a *Api) GetConfig() *configuration.Configuration  { return a.Config }
func (a *Api) SetConfig(c *configuration.Configuration) { a.Config = c }

func (a *Api) GetKeys() *keys.Keys  { return a.Keys }
func (a *Api) SetKeys(k *keys.Keys) { a.Keys = k }

func (a *Api) GetDnsCache() *dns.Records  { return a.DnsCache }
func (a *Api) SetDnsCache(d *dns.Records) { a.DnsCache = d }

func (a *Api) GetWss() *wss.WebSockets  { return a.Wss }
func (a *Api) SetWss(w *wss.WebSockets) { a.Wss = w }

func (a *Api) GetConfChangeC() chan raftpb.ConfChange  { return a.confChangeC }
func (a *Api) SetConfChangeC(c chan raftpb.ConfChange) { a.confChangeC = c }

func (a *Api) GetCluster() *cluster.Cluster  { return a.Cluster }
func (a *Api) SetCluster(c *cluster.Cluster) { a.Cluster = c }

func (a *Api) GetReplication() *distributed.Replication  { return a.Replication }
func (a *Api) SetReplication(r *distributed.Replication) { a.Replication = r }

func (a *Api) GetKinds() *relations.RelationRegistry  { return a.Kinds }
func (a *Api) SetKinds(k *relations.RelationRegistry) { a.Kinds = k }

func (a *Api) GetKindsRegistry() map[string]ikinds.Kind  { return a.KindsRegistry }
func (a *Api) SetKindsRegistry(m map[string]ikinds.Kind) { a.KindsRegistry = m }

func (a *Api) GetManager() *manager.Manager  { return a.Manager }
func (a *Api) SetManager(m *manager.Manager) { a.Manager = m }

func (a *Api) GetVersion() *version.Version  { return a.Version }
func (a *Api) SetVersion(v *version.Version) { a.Version = v }
