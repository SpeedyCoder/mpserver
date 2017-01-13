package mpserver

import (
	"time"
)

// Add a new component after addTimeout and remove a Component after removeTimeout
func LoadBalancingComponent(addTimeout, removeTimeout time.Duration, c Component) Component{
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
	        			l := len(chans)
	        			if (l > 1) {
	        				last := chans[l-1]
	        				chans = chans[:l-1]
	        				close(last)
	        			}
	        			continue
	        		}
	        	}
        	}
        	
        	select {
        		case toWorkers <- val: { written = true }
        		case <- time.After(addTimeout): {
        			// Add a worker
        			chans = append(chans, make(chan bool, 1))
        			go worker(c, out, chans[len(chans)-1])
        		}
        	} 
        }
    }
}


