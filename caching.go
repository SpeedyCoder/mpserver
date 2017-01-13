package mpserver

import (
	"net/http"
	"log"
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

func CacheComponent(c Component) Component {
	return func (in, out ValueChan) {
		toWorker := make(ValueChan)
		fromWorker := make(ValueChan)
		go c(toWorker, fromWorker)

		cache := make(map[string](Any)

	    for val := range in {
	    	if (stringInSlice(val.Request.Method, CachableMethods)) {
	    		key := requestToString(val.Request)
		        elem, in := cache[key]

		        if (in) {
		            log.Println("In cache")
		            log.Println(elem)
		            val.Result = elem
		            out <- val
		        } else {
		            // Not in cache
		            toWorker <- val
		            res := <- fromWorker
		            cache[key] = res.Result
		            out <- res
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
