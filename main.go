package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/raft"
	"github.com/heyyakash/infinitystore/datastore"
	"github.com/heyyakash/infinitystore/helper"
	"github.com/heyyakash/infinitystore/models"
	raft_consensus "github.com/heyyakash/infinitystore/raft"
)

var Store *datastore.DataStore
var r *raft.Raft

func main() {
	nodeID := flag.String("node-id", "node1", "Node ID")
	raftAddr := flag.String("raft-addr", "localhost:8000", "Raft Server Address")
	httpAddr := flag.String("http-addr", ":8001", "Http server Address")
	flag.Parse()
	Store = Store.Create()
	router := mux.NewRouter()
	fsm := &raft_consensus.FSM{
		Store: Store,
	}
	dataDir := "data"
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create data directory: %s", err)
	}
	r = raft_consensus.SetupRaft(*nodeID, path.Join(dataDir, *nodeID), *raftAddr, fsm)
	router.HandleFunc("/set", setHandler).Methods("POST")
	router.HandleFunc("/get", getHandler).Methods("GET")
	router.HandleFunc("/delete", deleteHandler).Methods("DELETE")
	router.HandleFunc("/join", joinNodes).Methods("POST")
	log.Print("Server Up and Running")
	log.Fatal(http.ListenAndServe(*httpAddr, router))
}

func setHandler(w http.ResponseWriter, re *http.Request) {
	defer re.Body.Close()
	req, err := io.ReadAll(re.Body)
	if err != nil {
		helper.ResponseGenerator(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	log.Print("Decoded", req)
	// Store.SetValue(req.Key, req.Value)
	future := r.Apply(req, 500*time.Millisecond)
	if err := future.Error(); err != nil {
		helper.ResponseGenerator(w, "Couldn't Add to store", http.StatusInternalServerError)
		return
	}
	_ = future.Response()
	if err := future.Error(); err != nil {
		helper.ResponseGenerator(w, "Couldn't Add to store", http.StatusInternalServerError)
		return
	}
	helper.ResponseGenerator(w, "Set Successfull", http.StatusOK)
}

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
	helper.ResponseGenerator(w, models.SetRequest{
		Key:   key,
		Value: Store.GetValue(key),
	}, http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
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
	Store.DeleteValue(key)
	helper.ResponseGenerator(w, "Key Deleted", http.StatusOK)
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

	helper.ResponseGenerator(w, "Voted added", http.StatusOK)
}
