package mpserver

import "time"

// dynamicLoadBalance starts one worker using the startWorker 
// function. It the forwards values from the channel in to the 
// channel toWorkers, while the channel in is open. When in is 
// closed, the function shutdowns the workers by calling the
// shutdown function. If after addTimeout time none of the 
// workers read the value that balancer is trying to write, then
// the balancer starts a new worker if the current number of 
// workers is smaller than maxWorkers. Similarly, when no value
// is sent to the balancer for removeTimeout time, then it 
// shutdowns a worker if there is a more than one worker active.
func dynamicLoadBalance(in <-chan Value, toWorkers chan<- Value, 
        startWorker startFunc, shutdown shutdownFunc, 
        addTimeout, removeTimeout time.Duration, maxWorkers int){
    // Start the first worker
    shutdownChans := make([](chan bool), 1)
    shutdownChans[0] = make(chan bool, 1)
    go startWorker(shutdownChans[0])

    var val Value
    nWorkers := 1 // Current number of workers.
    written := true
    ok := true
    for ok {
        if (written) {
            // Read a value or remove a worker
            select {
                case val, ok = <- in: {
                    if (!ok) {
                        // In was close, so I need to shutdown 
                        // all workers.
                        shutdown(shutdownChans)
                        continue
                    }
                    written = false
                }
                case <- time.After(removeTimeout): {
                    // Remove a worker if possible.
                    if (nWorkers > 1) {
                        last := shutdownChans[nWorkers-1]
                        shutdownChans = 
                            shutdownChans[:nWorkers-1]
                        close(last)
                        nWorkers--
                    }
                    continue
                }
            }
        }
        
        // Try to write the current value or add a worker.
        select {
            case toWorkers <- val: { written = true }
            case <- time.After(addTimeout): {
                // Add a worker if possible.
                if (nWorkers < maxWorkers) {
                    shutdownChans = append(
                        shutdownChans, make(chan bool, 1))
                    go startWorker(shutdownChans[nWorkers])
                    nWorkers++
                }
            }
        } 
    }
}

// DynamicLoadBalancer returns a component that performs dynamic 
// load balancing. That is it starts one worker instance, that 
// can be safely shutdown. When a value is waiting to be 
// processed for addTimeout, the balancer starts a new worker 
// provided the current number of workers is smaller than 
// maxWorkers. A worker is shutdown if the balancer doesn't get
// any new value for removeTimeout provided there is more than
// one worker.
func DynamicLoadBalancer(component Component, maxWorkers int,
             addTimeout, removeTimeout time.Duration) Component {
    return func (in <-chan Value, out chan<- Value) {
        toWorkers := GetChan()

        dynamicLoadBalance(in, toWorkers, 
            startComponent(component, toWorkers, out), 
            shutdownComponents(out), addTimeout, 
            removeTimeout, maxWorkers)
    }
}

// DynamicLoadBalancerWriter returns a writer that performs 
// dynamic load balancing. That is it starts one worker instance, 
// that can be safely shutdown. When a value is waiting to be 
// processed for addTimeout, the balancer starts a new worker 
// provided the current number of workers is smaller than 
// maxWorkers. A worker is shutdown if the balancer doesn't get
// any new value for removeTimeout provided there is more than
// one worker.
func DynamicLoadBalancerWriter(writer Writer, maxWorkers int,
                      addTimeout, removeTimeout time.Duration) Writer {
    return func (in <-chan Value) {
        toWorkers := GetChan()

        dynamicLoadBalance(in, toWorkers, 
            startWriter(writer, toWorkers), 
            shutdownWriters, addTimeout, 
            removeTimeout, maxWorkers)
    }
}