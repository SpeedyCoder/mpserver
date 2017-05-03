package mpserver

import (
	"github.com/orcaman/concurrent-map"
	"time"
	"log"
)

type Storage interface {
	Get(string) (interface{}, bool)
	Set(string, interface{})
	Remove(string)
	CompareAndRemove(string, interface{}) bool
	Keys() []string
}

func NewMemStore() Storage {
	cMap := cmap.New()
	return &cMap
}

type StoreValue struct {
	Value interface{}
	Time time.Time
}

// Method that removes expired items from the store
func StoreCleaner(store Storage, shutDown <-chan bool, sleepTime time.Duration) {
	done := false
	for !done {
		time.Sleep(sleepTime)
		select {
			case <-shutDown: { done = true; continue }
			default: {}
		}
		for _, key := range store.Keys() {
			elem, _ := store.Get(key)
			storeValue := elem.(StoreValue)
			if (storeValue.Time.Before(time.Now())) {
				// Cache can be updated at this point, so the following
				// Remove can remove an entry, which hasn't expired yet
				if (store.CompareAndRemove(key, elem)) {
					log.Println("Item removed")
				}
			}
		}
	}
}