package cluster

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/logger"
	"strings"
)

func (cluster *Cluster) ListenObjects(agent string) {
	for {
		select {
		case data, ok := <-cluster.KVStore.ObjectsC:
			if ok {
				// Kind should be encoded in the key
				split := strings.Split(data.Key, ".")
				kind := split[0]

				response := SendRequest(cluster.Client, &authentication.User{Username: cluster.KVStore.Agent, Domain: "localhost:1443"}, fmt.Sprintf("https://localhost:1443/api/v1/apply/%s/%s", kind, data.Agent), data.Val)

				if !response.Success {
					if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
						logger.Log.Error(errors.New(response.ErrorExplanation).Error())
					}
				}
				break
			}
		}

	}
}
