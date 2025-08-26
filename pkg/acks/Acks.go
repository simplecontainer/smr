package acks

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/smaps"
)

var ACKS = New()

func New() *Acks {
	return &Acks{
		Acks:    smaps.New(),
		Timeout: configuration.Timeout.AcknowledgmentTimeout,
	}
}

func (acks *Acks) Ack(UUID uuid.UUID) error {
	ackChanTmp, ok := acks.Acks.Map.Load(UUID)
	if !ok {
		return nil
	}

	ackChanTmp.(chan bool) <- true
	return nil
}

// Wait waits for acknowledgment for the given UUID or times out after a duration.
func (acks *Acks) Wait(UUID uuid.UUID) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), acks.Timeout)
	defer cancel()

	ackChan := make(chan bool)

	acks.Acks.Map.Store(UUID, ackChan)

	defer func() {
		if ackChanTmp, ok := acks.Acks.Map.Load(UUID); ok {
			close(ackChanTmp.(chan bool))
			acks.Acks.Map.Delete(UUID)
		}
	}()

	select {
	case <-ctxTimeout.Done():
		return errors.New("acknowledgment timed out after 10 seconds")
	case <-ackChan:
		return nil
	}
}
