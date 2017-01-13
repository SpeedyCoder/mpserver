package mpserver

import (
	"net/http"
	"log"
	"time"
	"github.com/orcaman/concurrent-map"
)

/** TODO:
  * 	cache invalidation 
*/

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
    res := r.Method + r.URL.String()
    log.Println(res)
    return res
}

type cacheValue struct {
	value Any
	time time.Time
}

// Method that removes expired items from the cache
func cacheCleaner(cache cmap.ConcurrentMap, shutDown <-chan bool, sleepTime time.Duration) {
	done := false
	for !done {
		time.Sleep(sleepTime)
		select {
			case <-shutDown: { done = true; continue }
			default: {}
		}
		for _, key := range cache.Keys() {
			elem, _ := cache.Get(key)
			cv := elem.(cacheValue)
			if (cv.time.Before(time.Now())) {
				cache.Remove(key)
				log.Println("Item removed")
			}
		}
	}
}

func CacheComponent(c Component, expiration time.Duration) Component {
	return func (in <-chan Value, out chan<- Value) {
		toWorker := make(ValueChan)
		fromWorker := make(ValueChan)
		go c(toWorker, fromWorker)

		cleanerShutDown := make(chan bool, 1)
		cache := cmap.New()
		go cacheCleaner(cache, cleanerShutDown, expiration)

		var computeAndAdd = func (key string, val Value, now time.Time) {
			toWorker <- val
            res := <- fromWorker
            cache.Set(key, cacheValue{res.Result, now.Add(expiration)})
            out <- res
		}

	    for val := range in {
	    	if (stringInSlice(val.Request.Method, CachableMethods)) {
	    		key := requestToString(val.Request)
		        elem, in := cache.Get(key)
		        now := time.Now()

		        if (in) {
		        	cv := elem.(cacheValue)
		            if (cv.time.After(now)) {
		            	log.Println("In cache\n")
		            	val.Result = cv.value
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
	    close(toWorker) // To shut down the component c
	    close(out)	
	}
}
