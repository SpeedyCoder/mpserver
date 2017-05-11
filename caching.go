package mpserver

import (
	"net/http"
	"log"
	"time"
)

/** TODO:
  * 	- remove oldest elements when the cache is full
  *     - use different strategy to generate keys for caching
  * 	- use map as input so that the caching component can be used 
  *		  with the load balancing component (do the same for session manager)
  				- then there would be multiple map cleaners
  *	    - make use cleaner optional
  *	    - fix examples using caching and session managers
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
	// TODO: make more specific
    res := r.Method + r.URL.String()
    log.Println(res)
    return res
}

func CacheComponent(cache Storage, worker Component, expiration time.Duration) Component {
	return func (in <-chan Value, out chan<- Value) {
		toWorker := GetChan()
		fromWorker := GetChan()
		go worker(toWorker, fromWorker)

		var computeAndAdd = func (key string, val Value, now time.Time) {
			toWorker <- val
            res := <- fromWorker
            cache.Set(key, StorageValue{res.GetResult(), now.Add(expiration)})
            out <- res
		}

	    for val := range in {
	    	if (stringInSlice(val.GetRequest().Method, CachableMethods)) {
	    		key := requestToString(val.GetRequest())
		        storageValue, in := cache.Get(key)
		        now := time.Now()

		        if (in) {
		            if (storageValue.Time.After(now)) {
		            	log.Println("In cache\n")
		            	val.SetResult(storageValue.Value)
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
	    close(toWorker) // To shut down the worker
	    close(out)	
	}
}
