package raft

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/channels"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/node"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/raft/v3"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/wal"
	"go.etcd.io/etcd/server/v3/wal/walpb"

	"go.uber.org/zap"
)

type Commit struct {
	data       []string
	applyDoneC chan<- struct{}
}

type RaftNode struct {
	proposeC    <-chan string            // proposed messages (k,v)
	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	nodeUpdate  chan node.Node
	commitC     chan<- *Commit // entries committed to log (k,v)
	errorC      chan<- error   // errors from raft session

	id          int         // client ID for raft session
	Peers       *node.Nodes // raft peer URLs
	join        bool        // node is joining an existing cluster
	waldir      string      // path to WAL directory
	snapdir     string      // path to snapshot directory
	getSnapshot func() ([]byte, error)

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	node                raft.Node
	IsLeader            atomic.Bool
	raftStorage         *raft.MemoryStorage
	wal                 *wal.WAL
	lastReplayedIndex   uint64
	firstReadyProcessed bool
	isRestart           bool
	started             time.Time

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	snapCount uint64
	transport *rafthttp.Transport
	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown
	httpdonec chan struct{} // signals http server shutdown complete

	TLSConfig *tls.Config

	logger *zap.Logger
}

var defaultSnapshotCount uint64 = 10000

// newRaftNode initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func NewRaftNode(keys *keys.Keys, TLSConfig *tls.Config, id uint64, peers *node.Nodes, join bool, getSnapshot func() ([]byte, error), channels *channels.Cluster) (*RaftNode, <-chan *Commit, <-chan error, <-chan *snap.Snapshotter) {
	commitC := make(chan *Commit)
	errorC := make(chan error)

	raftnode := &RaftNode{
		proposeC:    channels.Propose,
		confChangeC: channels.ConfChange,
		nodeUpdate:  channels.NodeUpdate,
		commitC:     commitC,
		errorC:      errorC,
		id:          int(id),
		Peers:       peers,
		join:        join,
		waldir:      fmt.Sprintf("/home/node/persistent/smr-%d", id),
		snapdir:     fmt.Sprintf("/home/node/persistent/smr-%d-snap", id),
		getSnapshot: getSnapshot,
		snapCount:   defaultSnapshotCount,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),

		started: time.Now(),

		TLSConfig: TLSConfig,

		logger: logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{"stdout"}, []string{"stderr"}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
	}

	go raftnode.startRaft(keys, TLSConfig)

	return raftnode, commitC, errorC, raftnode.snapshotterReady
}

func (rc *RaftNode) saveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index:     snap.Metadata.Index,
		Term:      snap.Metadata.Term,
		ConfState: &snap.Metadata.ConfState,
	}
	// save the snapshot file before writing the snapshot to the wal.
	// This makes it possible for the snapshot file to become orphaned, but prevents
	// a WAL snapshot entry from having no corresponding snapshot file.
	if err := rc.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	if err := rc.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	return rc.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (rc *RaftNode) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}
	firstIdx := ents[0].Index
	if firstIdx > rc.appliedIndex+1 {
		log.Fatalf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, rc.appliedIndex)
	}
	if rc.appliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[rc.appliedIndex-firstIdx+1:]
	}
	return nents
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rc *RaftNode) publishEntries(ents []raftpb.Entry) (<-chan struct{}, bool) {
	if len(ents) == 0 {
		return nil, true
	}

	data := make([]string, 0, len(ents))
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}
			s := string(ents[i].Data)
			data = append(data, s)
		case raftpb.EntryConfChange:
			if ents[i].Index <= rc.lastReplayedIndex {
				logger.Log.Info("ignoring ConfChange entry  (before last replayed index)", zap.Uint64("index", ents[i].Index), zap.Uint64("last replayed index", rc.lastReplayedIndex))
				continue
			}

			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)
			rc.confState = *rc.node.ApplyConfChange(cc)

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					n := node.NewNode()
					err := n.Parse(cc)

					if err != nil {
						log.Println("Invalid node configuration sent - conf change ignored.")
					} else {
						// Don't add itself

						if uint64(rc.id) != cc.NodeID {
							rc.transport.AddPeer(types.ID(cc.NodeID), []string{n.URL})
						}
					}
				}
			case raftpb.ConfChangeRemoveNode:
				n := node.NewNode()
				_ = n.Parse(cc)

				if cc.NodeID == uint64(rc.id) {
					return nil, false
				} else {
					// Update back to api.Nodes so that node can be removed
					rc.nodeUpdate <- *n
					rc.transport.RemovePeer(types.ID(cc.NodeID))
					rc.Peers.Remove(n)
				}
				break
			}
		}
	}

	var applyDoneC chan struct{}

	if len(data) > 0 {
		applyDoneC = make(chan struct{}, 1)
		select {
		case rc.commitC <- &Commit{data, applyDoneC}:
		case <-rc.stopc:
			return nil, false
		}
	}

	// after commit, update appliedIndex
	rc.appliedIndex = ents[len(ents)-1].Index

	return applyDoneC, true
}

func (rc *RaftNode) loadSnapshot() *raftpb.Snapshot {
	if wal.Exist(rc.waldir) {
		walSnaps, err := wal.ValidSnapshotEntries(rc.logger, rc.waldir)
		if err != nil {
			log.Fatalf("raft: error listing snapshots (%v)", err)
		}
		snapshot, err := rc.snapshotter.LoadNewestAvailable(walSnaps)
		if err != nil && !errors.Is(err, snap.ErrNoSnapshot) {
			log.Fatalf("raft: error loading snapshot (%v)", err)
		}
		return snapshot
	}
	return &raftpb.Snapshot{}
}

// openWAL returns a WAL ready for reading.
func (rc *RaftNode) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rc.waldir) {
		if err := os.Mkdir(rc.waldir, 0750); err != nil {
			log.Fatalf("raft: cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(zap.NewExample(), rc.waldir, nil)
		if err != nil {
			log.Fatalf("raft: create wal error (%v)", err)
		}
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	log.Printf("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	w, err := wal.Open(zap.NewExample(), rc.waldir, walsnap)
	if err != nil {
		log.Fatalf("raft: error loading wal (%v)", err)
	}

	return w
}

// replayWAL replays WAL entries into the raft instance.
func (rc *RaftNode) replayWAL() *wal.WAL {
	log.Printf("replaying WAL of member %d", rc.id)
	snapshot := rc.loadSnapshot()
	w := rc.openWAL(snapshot)
	_, st, ents, err := w.ReadAll()
	if err != nil {
		log.Fatalf("raft: failed to read WAL (%v)", err)
	}
	rc.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		rc.raftStorage.ApplySnapshot(*snapshot)
	}
	rc.raftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	rc.raftStorage.Append(ents)

	if len(ents) > 0 {
		rc.lastReplayedIndex = ents[len(ents)-1].Index
	}

	return w
}

func (rc *RaftNode) writeError(err error) {
	rc.stopHTTP()
	close(rc.commitC)
	rc.errorC <- err
	close(rc.errorC)
	rc.node.Stop()
	close(rc.stopc)
}

func (rc *RaftNode) startRaft(keys *keys.Keys, tlsConfig *tls.Config) {
	if !fileutil.Exist(rc.snapdir) {
		if err := os.Mkdir(rc.snapdir, 0750); err != nil {
			panic(fmt.Sprintf("smr: cannot create dir for snapshot (%v)", err))
		}
	}
	rc.snapshotter = snap.New(zap.NewExample(), rc.snapdir)

	oldwal := wal.Exist(rc.waldir)
	rc.wal = rc.replayWAL()

	// signal replay has finished
	rc.snapshotterReady <- rc.snapshotter

	rpeers := make([]raft.Peer, len(rc.Peers.ToString()))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	c := &raft.Config{
		ID:                        uint64(rc.id),
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   rc.raftStorage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}

	if oldwal || rc.join {
		rc.isRestart = true
		rc.node = raft.RestartNode(c)
	} else {
		rc.node = raft.StartNode(c, rpeers)
	}

	rc.transport = &rafthttp.Transport{
		Logger:             rc.logger,
		ID:                 types.ID(rc.id),
		ClusterID:          0x1000,
		Raft:               rc,
		DialTimeout:        5 * time.Second,
		DialRetryFrequency: rate.Every(300 * time.Millisecond),
		ServerStats:        stats.NewServerStats("", ""),
		LeaderStats:        stats.NewLeaderStats(zap.NewExample(), strconv.Itoa(rc.id)),
		ErrorC:             make(chan error, 1),
		TLSInfo: transport.TLSInfo{
			ClientCertAuth: true,
			KeyFile:        keys.Server.PrivateKeyPath,
			CertFile:       keys.Server.CertificatePath,
			TrustedCAFile:  keys.CA.CertificatePath,
			HandshakeFailure: func(conn *tls.Conn, err error) {
				fmt.Println(err.Error())
				conn.Close()
			},
		},
	}

	err := rc.transport.Start()
	if err != nil {
		panic(err)
	}

	for i := range rc.Peers.ToString() {
		if i+1 != rc.id {
			rc.transport.AddPeer(types.ID(i+1), []string{rc.Peers.ToString()[i]})
		}
	}

	go rc.serveRaft(keys, tlsConfig)
	go rc.serveChannels()
}

// stop closes http, closes all channels, and stops raft.
func (rc *RaftNode) stop() {
	rc.stopHTTP()
	close(rc.commitC)
	close(rc.errorC)

	rc.node.Stop()
	close(rc.stopc)
}

func (rc *RaftNode) stopHTTP() {
	rc.transport.Stop()
	close(rc.httpstopc)
	close(rc.httpdonec)
}

func (rc *RaftNode) Done() <-chan struct{} {
	return rc.stopc
}

func (rc *RaftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.Printf("publishing snapshot at index %d", rc.snapshotIndex)
	defer log.Printf("finished publishing snapshot at index %d", rc.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rc.appliedIndex {
		log.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, rc.appliedIndex)
	}
	rc.commitC <- nil

	rc.confState = snapshotToSave.Metadata.ConfState
	rc.snapshotIndex = snapshotToSave.Metadata.Index
	rc.appliedIndex = snapshotToSave.Metadata.Index
}

var snapshotCatchUpEntriesN uint64 = 100

func (rc *RaftNode) GetWALSize() (int64, error) {
	walFiles, err := os.ReadDir(rc.waldir)
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, file := range walFiles {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
		}
	}
	return totalSize, nil
}

func (rc *RaftNode) ForceSnapshot() error {
	data, err := rc.getSnapshot()
	if err != nil {
		return err
	}

	snap, err := rc.raftStorage.CreateSnapshot(rc.appliedIndex, &rc.confState, data)
	if err != nil {
		return err
	}

	if err := rc.saveSnap(snap); err != nil {
		return err
	}

	compactIndex := uint64(1)
	if rc.appliedIndex > 50 {
		compactIndex = rc.appliedIndex - 50
	}

	if err := rc.raftStorage.Compact(compactIndex); err != nil {
		if !errors.Is(err, raft.ErrCompacted) {
			return err
		}
	}

	rc.snapshotIndex = rc.appliedIndex
	return nil
}

func (rc *RaftNode) maybeTriggerSnapshot(applyDoneC <-chan struct{}) {
	if rc.appliedIndex-rc.snapshotIndex <= rc.snapCount {
		return
	}

	// wait until all committed entries are applied (or server is closed)
	if applyDoneC != nil {
		select {
		case <-applyDoneC:
		case <-rc.stopc:
			return
		}
	}

	log.Printf("start snapshot [applied index: %d | last snapshot index: %d]", rc.appliedIndex, rc.snapshotIndex)
	data, err := rc.getSnapshot()
	if err != nil {
		panic(err)
	}
	snap, err := rc.raftStorage.CreateSnapshot(rc.appliedIndex, &rc.confState, data)
	if err != nil {
		panic(err)
	}
	if err := rc.saveSnap(snap); err != nil {
		panic(err)
	}

	// More aggressive compaction - keep fewer entries
	compactIndex := uint64(1)
	if rc.appliedIndex > snapshotCatchUpEntriesN {
		compactIndex = rc.appliedIndex - snapshotCatchUpEntriesN
	}

	if err := rc.raftStorage.Compact(compactIndex); err != nil {
		if !errors.Is(err, raft.ErrCompacted) {
			panic(err)
		}
	} else {
		log.Printf("compacted log at index %d (kept only %d entries)", compactIndex, snapshotCatchUpEntriesN)
	}

	rc.snapshotIndex = rc.appliedIndex
}

func (rc *RaftNode) TransferLeadership(ctx context.Context, nodeID uint64) {
	rc.node.TransferLeadership(ctx, uint64(rc.id), nodeID)
}

func (rc *RaftNode) OnLeadershipChange(isLeader bool) {
	if isLeader {
		log.Printf("node %d is now the leader", rc.id)
	} else {
		log.Printf("node %d is no longer the leader", rc.id)
	}
}

func (rc *RaftNode) serveChannels() {
	snap, err := rc.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rc.confState = snap.Metadata.ConfState
	rc.snapshotIndex = snap.Metadata.Index
	rc.appliedIndex = snap.Metadata.Index

	defer rc.wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)

		for rc.proposeC != nil && rc.confChangeC != nil {
			select {
			case prop, ok := <-rc.proposeC:
				if !ok {
					rc.proposeC = nil
				} else {
					// blocks until accepted by raft state machine
					err = rc.node.Propose(context.TODO(), []byte(prop))

					if err != nil {
						logger.Log.Error(err.Error())
						return
					}
				}

			case cc, ok := <-rc.confChangeC:
				if !ok {
					rc.confChangeC = nil
				} else {
					confChangeCount++
					cc.ID = confChangeCount
					err = rc.node.ProposeConfChange(context.TODO(), cc)

					if err != nil {
						log.Panic(err)
					}
				}
			}
		}
		// client closed channel; shutdown raft if not already
		close(rc.stopc)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rc.node.Tick()

			// store raft entries to wal, then publish over commit channel
		case rd := <-rc.node.Ready():
			if !rc.firstReadyProcessed && rc.isRestart {
				rc.rebuildPeersFromCluster()
				rc.firstReadyProcessed = true
			}

			if rd.SoftState != nil {
				if rd.SoftState.RaftState == raft.StateLeader {
					if !rc.IsLeader.Load() {
						rc.OnLeadershipChange(true)
						rc.IsLeader.Store(true)
					}
				} else if rd.SoftState.RaftState == raft.StateFollower {
					if rc.IsLeader.Load() {
						log.Printf("node %d is no longer the leader", rc.id)
						rc.OnLeadershipChange(false)
						rc.IsLeader.Store(false)
					}
				}
			}

			// Must save the snapshot file and WAL snapshot entry before saving any other entries
			// or hardstate to ensure that recovery after a snapshot restore is possible.
			if !raft.IsEmptySnap(rd.Snapshot) {
				rc.saveSnap(rd.Snapshot)
			}

			rc.wal.Save(rd.HardState, rd.Entries)
			if !raft.IsEmptySnap(rd.Snapshot) {
				rc.raftStorage.ApplySnapshot(rd.Snapshot)
				rc.publishSnapshot(rd.Snapshot)
			}

			rc.raftStorage.Append(rd.Entries)
			rc.transport.Send(rc.processMessages(rd.Messages))

			if rd.SoftState != nil {
				rc.IsLeader.Store(rd.SoftState.RaftState == raft.StateLeader)
			}

			applyDoneC, ok := rc.publishEntries(rc.entriesToApply(rd.CommittedEntries))

			if !ok {
				rc.stop()
				return
			}

			rc.maybeTriggerSnapshot(applyDoneC)

			rc.node.Advance()

		case err := <-rc.transport.ErrorC:
			rc.writeError(err)
			return

		case <-rc.stopc:
			rc.stop()
			return
		}
	}
}

// When there is a `raftpb.EntryConfChange` after creating the snapshot,
// then the confState included in the snapshot is out of date. so We need
// to update the confState before sending a snapshot to a follower.
func (rc *RaftNode) processMessages(ms []raftpb.Message) []raftpb.Message {
	for i := 0; i < len(ms); i++ {
		if ms[i].Type == raftpb.MsgSnap {
			ms[i].Snapshot.Metadata.ConfState = rc.confState
		}
	}
	return ms
}

func (rc *RaftNode) serveRaft(keys *keys.Keys, tlsConfig *tls.Config) error {
	logger.Log.Info(fmt.Sprintf("Starting raft listener at %s", rc.Peers.ToString()[rc.id-1]))

	url, err := url.Parse(rc.Peers.ToString()[rc.id-1])
	if err != nil {
		log.Fatalf("raft: Failed parsing URL (%v)", err)
	}

	ln, err := newStoppableListener(url, rc.httpstopc)
	if err != nil {
		log.Fatalf("raft: Failed to listen rafthttp (%v)", err)
	}

	server := &http.Server{
		Handler:   rc.transport.Handler(),
		TLSConfig: tlsConfig,
	}

	err = server.ServeTLS(ln, "", "")

	if err != nil {
		return err
	}

	select {
	case <-rc.httpdonec:
		return nil
	default:
		log.Fatalf("raft: Failed to serve rafthttp (%v)", err)
	}

	return nil
}

func (rc *RaftNode) Process(ctx context.Context, m raftpb.Message) error {
	return rc.node.Step(ctx, m)
}
func (rc *RaftNode) IsIDRemoved(_ uint64) bool { return false }
func (rc *RaftNode) ReportUnreachable(id uint64) {
	rc.node.ReportUnreachable(id)
}
func (rc *RaftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	rc.node.ReportSnapshot(id, status)
}

func (rc *RaftNode) rebuildPeersFromCluster() {
	rc.logger.Info("rebuilding peers from the configuration", zap.String("cluster", strings.Join(rc.Peers.ToString(), ", ")))

	for _, peer := range rc.Peers.Nodes {
		if peer.NodeID != uint64(rc.id) {
			rc.logger.Info("adding node as peer", zap.Uint64("node", peer.NodeID))

			rc.node.ApplyConfChange(raftpb.ConfChange{
				Type:   raftpb.ConfChangeAddNode,
				ID:     peer.NodeID,
				NodeID: peer.NodeID,
			})
		}
	}
}
