package acks

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/smaps"
	"time"
)

var ACKS = New()

func New() *Acks {
	return &Acks{Acks: smaps.New()}
}

func (acks *Acks) Ack(UUID uuid.UUID) error {
	// Be CAUTIOUS when using wait since it is blocking till ack received or timeout
	ackChanTmp, ok := acks.Acks.Map.Load(UUID)

	if ok {
		ackChanTmp.(chan bool) <- true
	}

	return nil
}

func (acks *Acks) Wait(UUID uuid.UUID) error {
	// Be CAUTIOUS when using wait since it is blocking till ack received or timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var ackChan = make(chan bool)
	acks.Acks.Map.Store(UUID, ackChan)

	for {
		select {
		case <-ctxTimeout.Done():
			ackChanTmp, ok := acks.Acks.Map.Load(UUID)

			if ok {
				close(ackChanTmp.(chan bool))
				acks.Acks.Map.Delete(UUID)
			}

			return errors.New("object apply takes more than usual - wait timeout")
		case <-ackChan:
			ackChanTmp, ok := acks.Acks.Map.Load(UUID)

			if ok {
				close(ackChanTmp.(chan bool))
				acks.Acks.Map.Delete(UUID)
			}

			return nil
		}
	}
}
