package mpserver

import (
    "log"
    "time"
)

// Function for starting a worker
type startFunc func(chan bool)
// Function for shutting down an array of workers
type shutdownFunc func([](chan bool))

func staticLoadBalance(in <-chan Value, toWorkers ValueChan,
          startWorker startFunc, shutdown shutdownFunc, nWorkers int) {
    shutdownChans := make([](chan bool), nWorkers)
    for i := 0; i < nWorkers; i++ {
        shutdownChans[i] = make(chan bool, 1)
        go startWorker(shutdownChans[i])
    }
    for val := range in {
        toWorkers <- val
    }
    shutdown(shutdownChans)
}

func startComponent(comp Component, in <-chan Value, 
                                    out chan<- Value) startFunc {
    return  func(end chan bool) {
        toComponent := make(ValueChan)
        fromComponent := make(ValueChan)
        go comp(toComponent, fromComponent)

        done := false
        for !done {
            select {
                case val := <- in: {
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
}

func shutdownComponents(out chan<- Value) shutdownFunc {
    return func (shutdownChans [](chan bool)) {
        // Shut down the workers
        for _, schan := range shutdownChans {
            schan <- true
            // The second write is done only after
            // the worker reads the first value
            schan <- true
        }
        // All workers have now terminated, so 
        // I can close the out channel and 
        // propagate the termination
        close(out)
    }
}

func startWriter(writer Writer, toWorkers <-chan Value) startFunc {
    return  func(end chan bool) {
        toWriter:= make(ValueChan)
        go writer(toWriter)

        done := false
        for !done {
            select {
                case val := <- toWorkers: {
                    toWriter <- val
                }
                case <- end: {
                    // Closing the channel to the Writer 
                    // will shut down the Writer
                    close(toWriter)
                    done = true
                }
            }
        }
    }
}

func shutdownWriters(shutdownChans [](chan bool)) {
    // Shut down the workers
    for _, schan := range shutdownChans {
        close(schan)
    }
}

func StaticLoadBalancer(worker Component, nWorkers int) Component {
    return func (in <-chan Value, out chan<- Value) {
        toWorkers := make(ValueChan)
        staticLoadBalance(in, toWorkers, 
            startComponent(worker, toWorkers, out), 
            shutdownComponents(out), nWorkers)
    }
}

func StaticLoadBalancerWriter(worker Writer, nWorkers int) Writer {
    return func (in <-chan Value) {
        toWorkers := make(ValueChan)
        staticLoadBalance(in, toWorkers, 
            startWriter(worker, toWorkers), 
            shutdownWriters, nWorkers)
    }
}

func dynamicLoadBalance(in <-chan Value, toWorkers ValueChan, 
                        startWorker startFunc, shutdown shutdownFunc, 
                        addTimeout, removeTimeout time.Duration, 
                        maxWorkers int) {
    // Start the first worker
    shutdownChans := make([](chan bool), 1)
    shutdownChans[0] = make(chan bool, 1)
    go startWorker(shutdownChans[0])

    var val Value
    nWorkers := 1
    written := true
    ok := true
    for ok {
        if (written) {
            select {
                case val, ok = <- in: {
                    if (!ok) {
                        // Shutdown all workers
                        shutdown(shutdownChans)
                        continue
                    }
                    written = false
                }
                case <- time.After(removeTimeout): {
                    // Remove a worker
                    if (nWorkers > 1) {
                        last := shutdownChans[nWorkers-1]
                        shutdownChans = shutdownChans[:nWorkers-1]
                        close(last)
                        nWorkers--
                        log.Println("Worker removed.")
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
                    shutdownChans = append(shutdownChans, make(chan bool, 1))
                    go startWorker(shutdownChans[nWorkers])
                    nWorkers++
                    log.Println("Worker added.")
                }
            }
        } 
    }
}

// Add a new component after addTimeout and remove a Component after removeTimeout
func DynamicLoadBalancer(component Component, maxWorkers int,
                   addTimeout, removeTimeout time.Duration) Component {
    return func (in <-chan Value, out chan<- Value) {
        toWorkers := make(ValueChan)

        dynamicLoadBalance(in, toWorkers, 
            startComponent(component, toWorkers, out), 
            shutdownComponents(out), addTimeout, 
            removeTimeout, maxWorkers)
    }
}

// Add a new worker after addTimeout and remove a Component after removeTimeout
func DynamicLoadBalancerWriter(writer Writer, maxWorkers int,
                      addTimeout, removeTimeout time.Duration) Writer {
    return func (in <-chan Value) {
        toWorkers := make(ValueChan)

        dynamicLoadBalance(in, toWorkers, 
            startWriter(writer, toWorkers), 
            shutdownWriters, addTimeout, 
            removeTimeout, maxWorkers)
    }
}


