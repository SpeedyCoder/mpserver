package mpserver

import (
	"github.com/SpeedyCoder/concurrent-map"
	"time"
)

// StorageValue is a type a values that are stored in the Storage
// object.
type StorageValue struct {
	Value interface{} // The stored value itself.
	Time time.Time 	  // Expiration time for this value.
}

// Storage represents a thread safe type that represents a 
// mapping from string keys to StorageValues.
type Storage interface {
	// Get returns the storage value for the key and a boolean 
	// indicating whether the key is stored in the map.
	Get(key string) (StorageValue, bool)

	// Set stores the mapping from the given key to the given 
	// value.
	Set(key string, value StorageValue)

	// Remove removes the key-value pair from the map for the 
	// provided key.
	Remove(key string)

	// CompareAndRemove checks if the value stored in the map for
	// the given key is deep equal to the provided value. If that
	// is the case it removes the key-value pair from the map.
	// The boolean returned indicates whether the key-value pair  
	// was removed.
	CompareAndRemove(key string, value StorageValue) bool
	
	// Keys returns a slice that contains all keys that are 
	// stored in the mapping.
	Keys() []string
}

// memStorage is a type that implements the Storage interface.
// It uses sharded map internally to store the mapping.
type memStorage struct {
	cMap *cmap.ConcurrentMap
}

func (ms memStorage) Get(key string) (StorageValue, bool) {
	elem, in := ms.cMap.Get(key)
	if (in) {
		storageValue := elem.(StorageValue)
		return storageValue, true
	}

	return StorageValue{}, false
}

func (ms memStorage) Set(key string, value StorageValue) {
	ms.cMap.Set(key, value)
}

func (ms memStorage) Remove(key string) {
	ms.cMap.Remove(key)
}

func (ms memStorage) CompareAndRemove(key string, 
									  value StorageValue) bool {
	return ms.cMap.CompareAndRemove(key, value)
}

func (ms memStorage) Keys() []string {
	return ms.cMap.Keys()
}

// NewMemStorage returns a Storage object that stores the values
// in memory. 
func NewMemStorage() Storage {
	cMap := cmap.New()
	return memStorage{&cMap}
}


// StorageCleaner is a function that repeatedly removes expired 
// values from the provided Storage object. After it goes through
// all keys in the mapping it sleeps for time specified by the
// sleepTime argument. The function can be shutdown, by sending a
// message on the shutdown channel.
func StorageCleaner(storage Storage, shutDown <-chan bool, 
					sleepTime time.Duration) {
	done := false
	for !done {
		time.Sleep(sleepTime)
		select {
			case <-shutDown: { done = true; continue }
			default: {}
		}
		for _, key := range storage.Keys() {
			storageValue, _ := storage.Get(key)
			if (storageValue.Time.Before(time.Now())) {
				storage.CompareAndRemove(key, storageValue)
			}
		}
	}
}