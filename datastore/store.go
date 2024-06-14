package datastore

import "sync"

type DataStore struct {
	Store map[string]string
	mutex sync.RWMutex
}

func (d *DataStore) Create() *DataStore {
	return &DataStore{
		Store: make(map[string]string),
	}
}

func (d *DataStore) SetValue(key string, value string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Store[key] = value
}

func (d *DataStore) GetValue(key string) string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.Store[key]
}

func (d *DataStore) DeleteValue(key string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.Store, key)
}

func (d *DataStore) KeyExists(key string) bool {
	_, exists := d.Store[key]
	return exists
}

func (d *DataStore) GetAll() map[string]string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.Store
}
