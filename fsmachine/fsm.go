package fsmachine

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/hashicorp/raft"
	"github.com/heyyakash/infinitystore/datastore"
	"github.com/heyyakash/infinitystore/models"
)

type FSM struct {
	Store *datastore.DataStore
}

// Apply function
func (f *FSM) Apply(log *raft.Log) interface{} {
	var req models.Request
	if err := json.Unmarshal(log.Data, &req); err != nil {
		return err
	}
	switch req.Action {
	case "set":
		f.Store.SetValue(req.Key, req.Value)
	case "delete":
		f.Store.DeleteValue(req.Key)
	default:
		return errors.New("unknown action: " + req.Action)
	}
	return nil
}

type fsmSnapshot struct {
	state map[string]string
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	// Encode state json
	data, err := json.Marshal(s.state)
	if err != nil {
		sink.Cancel()
		return err
	}

	// write state to sink
	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *fsmSnapshot) Release() {}

// Snapshot function
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	state := f.Store.GetAll()

	return &fsmSnapshot{state: state}, nil
}

// Restore function
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
