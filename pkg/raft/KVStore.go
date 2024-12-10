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
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"log"
	"strings"
	"sync"

	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
)

// a key-value store backed by raft
type KVStore struct {
	proposeC    chan<- string // channel for proposing updates
	EtcdC       chan KV
	ObjectsC    chan KV
	ConfChangeC chan<- raftpb.ConfChange // channel for proposing updates
	Agent       string
	mu          sync.RWMutex
	client      *client.Http
	kvStore     *badger.DB
	snapshotter *snap.Snapshotter
}

type KV struct {
	Key      string
	Val      string
	Category string
	Agent    string
}

func NewKVStore(snapshotter *snap.Snapshotter, badger *badger.DB, client *client.Http, proposeC chan<- string, commitC <-chan *Commit, errorC <-chan error, etcdC chan KV, objectsC chan KV) *KVStore {
	s := &KVStore{proposeC: proposeC, EtcdC: etcdC, ObjectsC: objectsC, kvStore: badger, client: client, snapshotter: snapshotter}
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

	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s
}

func (s *KVStore) Propose(k string, v string, agent string) {
	var buf strings.Builder

	category := static.CATEGORY_PLAIN

	if strings.HasSuffix(k, ".object") {
		category = static.CATEGORY_OBJECT
	}

	if err := gob.NewEncoder(&buf).Encode(KV{k, v, category, agent}); err != nil {
		log.Fatal(err)
	}

	s.proposeC <- buf.String()
}

func (s *KVStore) ProposeEtcd(k string, v string, agent string) {
	URL := fmt.Sprintf("https://%s/api/v1/database/get/%s", s.client.Clients["root"].API, k)
	response := objects.SendRequest(s.client.Clients["root"].Http, URL, "GET", nil)

	if response.Success {
		b64decoded, err := base64.StdEncoding.DecodeString(response.Data[k].(string))

		if err != nil {
			logger.Log.Error(err.Error())
		} else {
			if string(b64decoded) == v {
				return
			}
		}
	}

	var buf strings.Builder

	if err := gob.NewEncoder(&buf).Encode(KV{k, v, static.CATEGORY_ETCD, agent}); err != nil {
		log.Fatal(err)
	}

	s.proposeC <- buf.String()
}

func (s *KVStore) readCommits(commitC <-chan *Commit, errorC <-chan error) {
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

		for _, data := range commit.data {
			var dataKv KV

			dec := gob.NewDecoder(bytes.NewBufferString(data))
			if err := dec.Decode(&dataKv); err != nil {
				log.Fatalf("raftexample: could not decode message (%v)", err)
			}

			s.mu.Lock()

			switch dataKv.Category {
			case static.CATEGORY_OBJECT:
				if dataKv.Agent == s.Agent {
					URL := fmt.Sprintf("https://%s/api/v1/database/update/%s", s.client.Clients["root"].API, dataKv.Key)
					response := objects.SendRequest(s.client.Clients["root"].Http, URL, "PUT", []byte(dataKv.Val))

					logger.Log.Debug("distributed object update", zap.String("URL", URL), zap.String("data", dataKv.Val))

					if !response.Success {
						log.Panic(errors.New(response.ErrorExplanation))
					}
				} else {
					s.ObjectsC <- dataKv
				}
				break

			case static.CATEGORY_PLAIN:
				URL := fmt.Sprintf("https://%s/api/v1/database/update/%s", s.client.Clients["root"].API, dataKv.Key)
				response := objects.SendRequest(s.client.Clients["root"].Http, URL, "PUT", []byte(dataKv.Val))

				logger.Log.Debug("distributed object update", zap.String("URL", URL), zap.String("data", dataKv.Val))

				if !response.Success {
					log.Panic(errors.New(response.ErrorExplanation))
				}
				break
			case static.CATEGORY_ETCD:
				URL := fmt.Sprintf("https://%s/api/v1/database/update/%s", s.client.Clients["root"].API, dataKv.Key)
				response := objects.SendRequest(s.client.Clients["root"].Http, URL, "PUT", []byte(dataKv.Val))

				logger.Log.Debug("distributed object update", zap.String("URL", URL), zap.String("data", dataKv.Val))

				if !response.Success {
					log.Panic(errors.New(response.ErrorExplanation))
				}

				s.EtcdC <- dataKv
				break
			}

			s.mu.Unlock()
		}
		close(commit.applyDoneC)
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

func (s *KVStore) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.kvStore)
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
	var store map[string]string
	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}

	s.mu.Lock()

	for k, v := range store {
		URL := fmt.Sprintf("https://%s/api/v1/database/update/%s", s.client.Clients["root"].API, k)
		response := objects.SendRequest(s.client.Clients["root"].Http, URL, "PUT", []byte(v))

		logger.Log.Debug("distributed object update", zap.String("URL", URL), zap.String("data", v))

		if !response.Success {
			return errors.New(response.ErrorExplanation)
		}
	}

	s.mu.Unlock()

	return nil
}
