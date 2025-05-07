package mock

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	mock_platforms "github.com/simplecontainer/smr/pkg/kinds/containers/mock"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/static"
)

func GetShared(registryMock *mock_platforms.MockRegistry) *shared.Shared {
	keys := keys.NewKeys()

	keys.GenerateCA()
	keys.GenerateClient(configuration.NewDomains([]string{"testi.ng"}), configuration.NewIPs([]string{}), "node-1")

	node := node.NewNode()
	node.NodeID = 1
	node.NodeName = "node-1"

	httpClient, _ := clients.GenerateHttpClients("node-1", keys, configuration.HostPort{
		Host: "",
		Port: "1443",
	}, &cluster.Cluster{Node: node})

	sh := &shared.Shared{
		Registry: registryMock,
		User: &authentication.User{
			Username: "node-1",
			Domain:   "localhost",
		},
		Watchers: nil,
		DnsCache: dns.New("node-1", nil, nil),
		Manager:  &manager.Manager{Config: NewConfig(static.PLATFORM_MOCKER)},
		Client:   httpClient,
	}

	sh.Watchers = watcher.NewWatchers()

	return sh
}
