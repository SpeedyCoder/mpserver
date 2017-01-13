package mpserver

import (
	"net/http"
	"log"
	"time"
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

func CacheComponent(c Component, expiration time.Duration) Component {
	return func (in <-chan Value, out chan<- Value) {
		toWorker := make(ValueChan)
		fromWorker := make(ValueChan)
		go c(toWorker, fromWorker)

		cache := make(map[string]cacheValue)
		var computeAndAdd = func (key string, val Value, now time.Time) {
			toWorker <- val
            res := <- fromWorker
            cache[key] = cacheValue{res.Result, now.Add(expiration)}
            out <- res
		}

	    for val := range in {
	    	if (stringInSlice(val.Request.Method, CachableMethods)) {
	    		key := requestToString(val.Request)
		        elem, in := cache[key]
		        now := time.Now()

		        if (in) {
		            if (elem.time.After(now)) {
		            	log.Println("In cache\n")
		            	val.Result = elem.value
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
	    close(toWorker) // To shut down the component c
	    close(out)	
	}
}
