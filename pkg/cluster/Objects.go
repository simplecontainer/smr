package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
	"strings"
)

func (c *Cluster) ListenObjects(agent string) {
	for {
		select {
		case data, ok := <-c.KVStore.ObjectsC:
			if ok {
				if data.Agent != agent {
					if strings.Contains(data.Val, "container") {
						container := &v1.ContainerDefinition{}
						err := json.Unmarshal([]byte(data.Val), container)

						if err != nil {
							logger.Log.Error(err.Error())
						}

						bytes, err := container.ToJsonStringWithKind()

						if err != nil {
							logger.Log.Error(err.Error())
						}

						response := sendRequest(c.Client, &authentication.User{"root", ""}, "https://localhost:1443/api/v1/apply", bytes)

						if !response.Success {
							if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
								logger.Log.Error(errors.New(response.ErrorExplanation).Error())
							} else {
								logger.Log.Info(fmt.Sprintf(response.ErrorExplanation))
							}
						}
					} else {
						fmt.Println("no update")
					}
				}
			}
		}
	}
}
