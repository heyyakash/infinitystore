package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/raft"
	"github.com/heyyakash/infinitystore/datastore"
	"github.com/heyyakash/infinitystore/fsmachine"
	"github.com/heyyakash/infinitystore/helper"
	"github.com/heyyakash/infinitystore/models"
	raft_consensus "github.com/heyyakash/infinitystore/raft"
)

var Store *datastore.DataStore
var r *raft.Raft

var (
	nodeID   *string
	raftAddr *string
	httpAddr *string
)

// Parsing flags
func Init() {
	nodeID = flag.String("node-id", "node1", "Node ID")
	raftAddr = flag.String("raft-addr", "localhost:8000", "Raft Server Address")
	httpAddr = flag.String("http-addr", ":8001", "Http server Address")
	flag.Parse()
}

func main() {
	Init()

	// Initializing Store
	Store = Store.Create()

	// Initializing Router
	router := mux.NewRouter()

	// Initializing Finite State Machine
	fsm := &fsmachine.FSM{
		Store: Store,
	}

	// Creating dir
	dataDir := "data"
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create data directory: %s", err)
	}

	// Setting up raft
	r = raft_consensus.SetupRaft(*nodeID, path.Join(dataDir, *nodeID), *raftAddr, fsm)

	// Adding Routes
	router.HandleFunc("/set", setHandler).Methods("POST")
	router.HandleFunc("/get", getHandler).Methods("GET")
	router.HandleFunc("/join", joinNodes).Methods("POST")

	// Starting Server
	log.Print("Server Up and Running")
	log.Fatal(http.ListenAndServe(*httpAddr, router))
}

// handler to handle set and delete actions
func setHandler(w http.ResponseWriter, re *http.Request) {
	defer re.Body.Close()
	req, err := io.ReadAll(re.Body)
	if err != nil {
		helper.ResponseGenerator(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	future := r.Apply(req, 500*time.Millisecond)
	if err := future.Error(); err != nil {
		helper.ResponseGenerator(w, "Couldn't Add to store", http.StatusInternalServerError)
		return
	}
	response := future.Response()
	if err, ok := response.(error); ok {
		helper.ResponseGenerator(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	helper.ResponseGenerator(w, "Set Successfull", http.StatusOK)
}

// handler to handle get operation
func getHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys) == 0 {
		helper.ResponseGenerator(w, "Missing Key", http.StatusBadRequest)
		return
	}

	key := keys[0]
	if exists := Store.KeyExists(key); !exists {
		helper.ResponseGenerator(w, "Key does not exist", http.StatusNotFound)
		return
	}
	helper.ResponseGenerator(w, models.GetResponse{
		Key:   key,
		Value: Store.GetValue(key),
	}, http.StatusOK)
}

func joinNodes(w http.ResponseWriter, req *http.Request) {
	nodeid := req.URL.Query().Get("nodeid")
	nodeaddr := req.URL.Query().Get("nodeaddr")

	if r.State() != raft.Leader {
		helper.ResponseGenerator(w, "Not a leader", http.StatusBadRequest)
		return
	}

	if err := r.AddVoter(raft.ServerID(nodeid), raft.ServerAddress(nodeaddr), 0, 0).Error(); err != nil {
		helper.ResponseGenerator(w, "Failed to add voter ", http.StatusInternalServerError)
		return
	}

	helper.ResponseGenerator(w, "Voter added", http.StatusOK)
}
