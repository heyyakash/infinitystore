package raft_consensus

import (
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/heyyakash/infinitystore/fsmachine"
)

func SetupRaft(nodeID string, raftDir string, raftAddr string, fsm *fsmachine.FSM) *raft.Raft {
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
