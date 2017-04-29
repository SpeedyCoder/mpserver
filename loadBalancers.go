package mpserver

import (
	"time"
)

func StaticLoadBalancer(worker Component, n int) Component {
	return func (in <-chan Value, out chan<- Value) {
		for i := 0; i < n-1; i++ {
			go worker(in, out)
		}
		worker(in, out)
	}
}

// Add a new component after addTimeout and remove a Component after removeTimeout
func DynamicLoadBalancer(addTimeout, removeTimeout time.Duration, c Component, maxWorkers int) Component{
	toWorkers := make(ValueChan)

	var worker = func (c Component, out chan<- Value, end chan bool) {
		toComponent := make(ValueChan)
		fromComponent := make(ValueChan)
		go c(toComponent, fromComponent)

		done := false
		for !done {
			select {
				case val := <- toWorkers: {
					toComponent <- val
					res := <- fromComponent
					out <- res
				}
				case <- end: {
					<- end
					// Closing the channel to the Component 
					// will shut down the Component
					close(toComponent)
					done = true
				}
			}
		}
	}

    return func (in <-chan Value, out chan<- Value) {
    	chans := make([](chan bool), 1)
    	chans[0] = make(chan bool, 1)
    	go worker(c, out, chans[0])

    	var val Value
    	nWorkers := 1
    	written := true
    	ok := true
        for ok {
        	if (written) {
        		select {
	        		case val, ok = <- in: {
	        			if (!ok) {
	        				// Shut down the workers
	        				for _, bchan := range chans {
	        					bchan <- true
	        					// The second write is done only after
	        					// the worker reads the firs value
	        					bchan <- true
	        				}
	        				// All workers have now terminated, so 
	        				// I can close the out channel and 
	        				// propagate the termination
	        				close(out)
	        				continue
	        			}
	        			written = false
	        		}
	        		case <- time.After(removeTimeout): {
	        			// Remove a worker
	        			if (nWorkers > 1) {
	        				last := chans[nWorkers-1]
	        				chans = chans[:nWorkers-1]
	        				close(last)
	        				nWorkers--
	        			}
	        			continue
	        		}
	        	}
        	}
        	
        	select {
        		case toWorkers <- val: { written = true }
        		case <- time.After(addTimeout): {
        			// Add a worker
        			if (nWorkers < maxWorkers) {
	        			chans = append(chans, make(chan bool, 1))
	        			go worker(c, out, chans[len(chans)-1])
	        			nWorkers++   				
        			}
        		}
        	} 
        }
    }
}


