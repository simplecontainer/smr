package network

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/kinds/network/implementation"
	"time"
)

func (network *Network) recreateNetworkWithRetry(networkObj *implementation.Network) error {
	for {
		select {
		case <-time.After(5 * time.Second):
			return errors.New("network didn't delete properly")
		case <-time.Tick(500 * time.Millisecond):
			if err := networkObj.Create(); err == nil {
				return nil
			}
		}
	}
}
