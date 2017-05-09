package mpserver

import (
	"github.com/orcaman/concurrent-map"
	"time"
	"log"
)

type StorageValue struct {
	Value interface{}
	Time time.Time
}

type Storage interface {
	Get(string) (StorageValue, bool)
	Set(string, StorageValue)
	Remove(string)
	CompareAndRemove(string, StorageValue) bool
	Keys() []string
}

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

func (ms memStorage) CompareAndRemove(key string, value StorageValue) bool {
	return ms.cMap.CompareAndRemove(key, value)
}

func (ms memStorage) Keys() []string {
	return ms.cMap.Keys()
}

func NewMemStorage() Storage {
	cMap := cmap.New()
	return memStorage{&cMap}
}


// Method that removes expired items from the storage
func StorageCleaner(storage Storage, shutDown <-chan bool, sleepTime time.Duration) {
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
				// Cache can be updated at this point, so the following
				// Remove can remove an entry, which hasn't expired yet
				if (storage.CompareAndRemove(key, storageValue)) {
					log.Println("Item removed")
				}
			}
		}
	}
}