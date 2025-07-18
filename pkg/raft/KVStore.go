// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/node"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"log"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
)

// a key-value store backed by raft
type KVStore struct {
	proposeC    chan<- string // channel for proposing updates
	DataC       chan KV.KV
	InSyncC     chan bool
	Replay      bool
	ConfChangeC chan<- raftpb.ConfChange // channel for proposing updates
	Node        *node.Node
	mu          sync.RWMutex
	snapshotter *snap.Snapshotter

	etcdClient    *clientv3.Client
	CommittedKeys sync.Map
}

func NewKVStore(etcdclient *clientv3.Client, snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *Commit, errorC <-chan error, dataC chan KV.KV, insyncC chan bool, join bool, replay bool, node *node.Node) (*KVStore, error) {
	s := &KVStore{
		etcdClient:    etcdclient,
		CommittedKeys: sync.Map{},
		proposeC:      proposeC,
		DataC:         dataC,
		InSyncC:       insyncC,
		snapshotter:   snapshotter,
		Replay:        replay,
		Node:          node,
	}

	snapshot, err := s.loadSnapshot()

	if err != nil {
		return nil, err
	}

	if snapshot != nil {
		log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)

		if err = s.recoverFromSnapshot(snapshot.Data); err != nil {
			return nil, err
		}
	}

	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s, nil
}

func (s *KVStore) Propose(k string, v []byte, node uint64) {
	var buf strings.Builder

	if err := gob.NewEncoder(&buf).Encode(KV.NewEncode(k, v, node)); err != nil {
		log.Fatal(err)
	}

	s.proposeC <- buf.String()
}

func (s *KVStore) readCommits(commitC <-chan *Commit, errorC <-chan error) {
	if s.Node.Accepting() {
		for commit := range commitC {
			if commit == nil {
				// signaled to load snapshot
				snapshot, err := s.loadSnapshot()

				if err != nil {
					log.Panic(err)
				}

				if snapshot != nil {
					log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
					if err = s.recoverFromSnapshot(snapshot.Data); err != nil {
						log.Panic(err)
					}
				}
				continue
			}

			s.mu.Lock()

			for _, data := range commit.data {
				s.DataC <- KV.NewDecode(gob.NewDecoder(bytes.NewBufferString(data)), s.Node.NodeID)
			}

			if s.Replay {
				s.Replay = false
				s.InSyncC <- true
			}

			s.mu.Unlock()

			close(commit.applyDoneC)
		}
		if err, ok := <-errorC; ok {
			log.Fatal(err)
		}
	} else {
		logger.Log.Info("node not accepting new objects")
	}
}

func (s *KVStore) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	references := make(map[string]string)

	s.CommittedKeys.Range(func(key, value any) bool {
		k := key.(string)
		references[k] = fmt.Sprintf("ref::%s", k)
		return true
	})

	return json.Marshal(references)
}

func (s *KVStore) loadSnapshot() (*raftpb.Snapshot, error) {
	snapshot, err := s.snapshotter.Load()
	if errors.Is(err, snap.ErrNoSnapshot) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *KVStore) recoverFromSnapshot(snapshot []byte) error {
	var references map[string]string
	if err := json.Unmarshal(snapshot, &references); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ref := range references {
		if !strings.HasPrefix(ref, "ref::") {
			continue
		}

		actualKey := strings.TrimPrefix(ref, "ref::")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := s.etcdClient.Get(ctx, actualKey)
		cancel()

		if err != nil {
			logger.Log.Info("failed to resolve reference from etcd", zap.String("reference", actualKey), zap.Error(err))
			continue
		}

		if len(resp.Kvs) == 0 {
			logger.Log.Info("reference not found in etcd", zap.String("reference", actualKey))
			continue
		}

		s.DataC <- KV.NewDecode(gob.NewDecoder(bytes.NewBuffer(resp.Kvs[0].Value)), s.Node.NodeID)
		s.CommittedKeys.Store(actualKey, true)
	}

	return nil
}

func (s *KVStore) Close() error {
	if s.etcdClient != nil {
		return s.etcdClient.Close()
	}
	return nil
}
