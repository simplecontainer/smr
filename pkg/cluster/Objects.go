package cluster

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/logger"
	"strings"
)

func (c *Cluster) ListenObjects(agent string) {
	for {
		select {
		case data, ok := <-c.KVStore.ObjectsC:
			if ok {
				// Kind is encoded in the key
				split := strings.Split(data.Key, ".")
				kind := split[0]

				response := SendRequest(c.Client, &authentication.User{"root", ""}, fmt.Sprintf("https://localhost:1443/api/v1/apply/%s", kind), data.Val)

				if !response.Success {
					if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
						logger.Log.Error(errors.New(response.ErrorExplanation).Error())
					} else {
						logger.Log.Info(fmt.Sprintf(response.ErrorExplanation))
					}
				}
			}
		}

	}
}
