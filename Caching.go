package mpserver

import (
	"net/http"
	"log"
	"time"
	"strings"
)

// stringInSlice checks if a provided slice contains the provided
// string.
func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

// HTTP methods that should be cached.
var CachableMethods = []string{"GET", "HEAD", "POST", "PATCH"}

// isCachable checks if the provided method is in the slice of 
// cachable methods.
func isCachable(method string) bool {
	return stringInSlice(method, CachableMethods)
}

// requestToString converts a request to string.
func requestToString(r *http.Request) string {
    res := r.Method + r.URL.String() + "HEADERS:"
    for key, value := range r.Header {
        res += key + ":" + strings.Join(value, "") + ";"
    }
    log.Println(res)
    return res
}

// CacheComponent generates a component that caches the generated
// result for all input jobs. That is if a job containing a 
// request that the component haven't seen been before arrives, 
// it forwards it to the worker and stores the result. If a 
// previously seen request arrives the component just uses the 
// stored result. The values expire after the specified time. 
// Then the result needs to be computed again.
func CacheComponent(cache Storage, worker Component, 
					expiration time.Duration) Component {
	return func (in <-chan Job, out chan<- Job) {
		toWorker := GetChan()
		fromWorker := GetChan()
		// Start the worker.
		go worker(toWorker, fromWorker)

		// Compute result and store it in the storage object.
		var computeAndStore = func (key string, job Job, 
									now time.Time) {
			toWorker <- job
            res := <- fromWorker
            // Store the result in the cache.
            cache.Set(key, StorageValue{
            	res.GetResult(), now.Add(expiration)})
            out <- res
		}

	    for job := range in {
	    	// Check if the HTTP method used is cachable.
	    	if (isCachable(job.GetRequest().Method)) {
	    		key := requestToString(job.GetRequest())
		        storageValue, in := cache.Get(key)
		        now := time.Now()

		        if (in) {
		        	// The result for this request is in 
		        	// the cache.
		            if (storageValue.Time.After(now)) {
		            	// The cached value hasn't expired yet,
		            	// so we can use it.
		            	job.SetResult(storageValue.Value)
		            	out <- job
		            } else {
		            	// The cached value has expired, so we 
		            	// need to compute the result again.
		            	computeAndStore(key, job, now)
		            }
		        } else {
		        	// The result for this request is not in 
		        	// the cache.
		            computeAndStore(key, job, now)
		        }
	    	} else {
	    		// Method is not cachable.
	    		toWorker <- job
	    		res := <- fromWorker
	    		out <- res
	    	}
	        
	    }
	    close(toWorker) // Shut down the worker.
	    close(out)
	}
}