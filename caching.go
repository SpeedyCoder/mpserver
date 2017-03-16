package mpserver

import (
	"net/http"
	"log"
	"time"
	"github.com/orcaman/concurrent-map"
)

/** TODO:
  * 	- remove oldest elements when the cache is full
  *     - use different strategy to generate keys for caching
  * 	- use map as input so that the caching component can be used 
  *		  with the load balancing component (do the same for session manager)
  				- then there would be multiple map cleaners
*/

type Store interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Remove(key string)
	Keys() []string
}

type cMapStore struct {
	cMap *cmap.ConcurrentMap
}

func (store *cMapStore) Get(key string) (interface{}, bool) {
	return store.cMap.Get(key)
}

func (store *cMapStore) Set(key string, value interface{}) {
	store.cMap.Set(key, value)
}

func (store *cMapStore) Remove(key string) {
	store.cMap.Remove(key)
}

func (store *cMapStore) Keys() []string {
	return store.cMap.Keys()
}

func NewMemStore() Store {
	cMap := cmap.New()
	cms := cMapStore{&cMap}
	return &cms
}


func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

var CachableMethods = []string{"GET", "HEAD", "POST", "PATCH"}

func requestToString(r *http.Request) string {
	// TODO: make more specific
    res := r.Method + r.URL.String()
    log.Println(res)
    return res
}

type MapValue struct {
	Value Any
	Time time.Time
}

// Method that removes expired items from the cache
func mapCleaner(cache Store, shutDown <-chan bool, sleepTime time.Duration) {
	done := false
	for !done {
		time.Sleep(sleepTime)
		select {
			case <-shutDown: { done = true; continue }
			default: {}
		}
		for _, key := range cache.Keys() {
			elem, _ := cache.Get(key)
			mapValue := elem.(MapValue)
			if (mapValue.Time.Before(time.Now())) {
				// Cache can be updated at this point, so the following
				// Remove can remove an entry, which hasn't expired yet
				cache.Remove(key)
				log.Println("Item removed")
			}
		}
	}
}

func CacheComponent(cache Store, worker Component, expiration time.Duration) Component {
	return func (in <-chan Value, out chan<- Value) {
		toWorker := make(ValueChan)
		fromWorker := make(ValueChan)
		go worker(toWorker, fromWorker)

		cleanerShutDown := make(chan bool, 1)
		go mapCleaner(cache, cleanerShutDown, expiration)

		var computeAndAdd = func (key string, val Value, now time.Time) {
			toWorker <- val
            res := <- fromWorker
            cache.Set(key, MapValue{res.Result, now.Add(expiration)})
            out <- res
		}

	    for val := range in {
	    	if (stringInSlice(val.Request.Method, CachableMethods)) {
	    		key := requestToString(val.Request)
		        elem, in := cache.Get(key)
		        now := time.Now()

		        if (in) {
		        	mapValue := elem.(MapValue)
		            if (mapValue.Time.After(now)) {
		            	log.Println("In cache\n")
		            	val.Result = mapValue.Value
		            	out <- val
		            } else {
		            	log.Println("Cache expired\n")
		            	computeAndAdd(key, val, now)
		            }
		        } else {
		        	log.Println("Not in cache\n")
		            computeAndAdd(key, val, now)
		        }
	    	} else {
	    		// Method is not Cachable
	    		toWorker <- val
	    		res := <- fromWorker
	    		out <- res
	    	}
	        
	    }
	    cleanerShutDown <- true
	    close(toWorker) // To shut down the worker
	    close(out)	
	}
}
