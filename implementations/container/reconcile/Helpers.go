package reconcile

import (
	"fmt"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"time"
)

func WaitForStop(containerObj *container.Container) {
	timeout := false
	waitForStop := make(chan string, 1)
	go func() {
		for {
			c, err := containerObj.Get()

			if err != nil {
				return
			}

			if timeout {
				return
			}

			if c != nil && c.State != "exited" {
				logger.Log.Info(fmt.Sprintf("waiting for container to exit %s", containerObj.Static.GeneratedName))
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		waitForStop <- "container exited proceed with delete for reconciliation"
	}()

	select {
	case res := <-waitForStop:
		logger.Log.Debug(fmt.Sprintf("%s %s", res, containerObj.Static.GeneratedName))
	case <-time.After(30 * time.Second):
		logger.Log.Debug("timed out waiting for the container to exit", zap.String("container", containerObj.Static.GeneratedName))
		timeout = true
	}
}
