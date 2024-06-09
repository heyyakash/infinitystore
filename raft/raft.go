package raft_consensus

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/heyyakash/infinitystore/datastore"
)

type FSM struct {
	Store *datastore.DataStore
}

func SetupRaft(nodeID string, raftDir string, raftAddr string, fsm *FSM) *raft.Raft {
	os.MkdirAll(raftDir, os.ModePerm)
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	address, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		log.Fatal("Couldn't resolve TCP address : ", err)
	}

	transport, err := raft.NewTCPTransport(raftAddr, address, 3, 10*time.Second, os.Stdout)
	if err != nil {
		log.Fatal("Failed to create transport : ", err)
	}

	boltStore, err := raftboltdb.NewBoltStore(path.Join(raftDir, "bolt"))
	if err != nil {
		log.Fatal("Failed to create new bolt store : ", err)
	}

	snapshot, err := raft.NewFileSnapshotStore(path.Join(raftDir, "snapshot"), 2, os.Stdout)
	if err != nil {
		log.Fatal("Failed to create new snapshot store : ", err)
	}

	logStore := boltStore
	r, err := raft.NewRaft(config, fsm, logStore, boltStore, snapshot, transport)
	if err != nil {
		log.Fatal("Failed to create raft : ", err)
	}

	r.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeID),
				Address: transport.LocalAddr(),
			},
		},
	})

	return r
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	var kvstore map[string]string
	if err := json.Unmarshal(log.Data, &kvstore); err != nil {
		panic(err)
	}
	fmt.Print(kvstore)
	for key, value := range kvstore {
		f.Store.SetValue(key, value)
	}
	return nil
}

type snapshotNoop struct{}

func (sn snapshotNoop) Persist(_ raft.SnapshotSink) error { return nil }
func (sn snapshotNoop) Release()                          {}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	// Implement snapshotting logic
	return snapshotNoop{}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	for key, _ := range f.Store.Store {
		f.Store.DeleteValue(key)
	}
	var kvstore map[string]string
	if err := json.NewDecoder(rc).Decode(&kvstore); err != nil {
		panic(err)
	}
	for key, value := range kvstore {
		f.Store.SetValue(key, value)
	}
	return rc.Close()
}
