package mpserver

// startFunc is a function used for starting a worker that can be
// shut down by sending a message on the provided signaling 
// channel.
type startFunc func(chan bool)

// shutdownFunc is a function used for shutting down an array of 
// workers represented by their shutdown channels.
type shutdownFunc func([](chan bool))

// staticLoadBalance starts nWorkers instances of workers using
// the startWorker function. It the forwards values from the 
// channel in to the channel toWorkers, while the channel in is
// open. When in is closed, the function shutdowns the workers
// by calling the shutdown function.
func staticLoadBalance(in <-chan Value, toWorkers chan<- Value,
                       startWorker startFunc, 
                       shutdown shutdownFunc, nWorkers int) {
    // Start the workers.
    shutdownChans := make([](chan bool), nWorkers)
    for i := 0; i < nWorkers; i++ {
        shutdownChans[i] = make(chan bool, 1)
        go startWorker(shutdownChans[i])
    }
    // Forward incoming values to the worker while the input 
    // channel is open.
    for val := range in {
        toWorkers <- val
    }
    // Shutdown the workers.
    shutdown(shutdownChans)
}

// startComponent returns a startFunc that starts a worker that 
// is represented by a component. 
//
// When the returned function is called it behaves as follows.
// It starts the worker and then forwards the values from the 
// channel in to the worker and waits for the result from the 
// worker, which it then sends to the out channel. When a signal 
// is sent on the end channel the worker is shutdown and the 
// function terminates.
func startComponent(worker Component, in <-chan Value, 
                    out chan<- Value) startFunc {
    return  func(end chan bool) {
        // Start the worker
        toComponent := GetChan()
        fromComponent := GetChan()
        go worker(toComponent, fromComponent)

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
                    // Closing the channel to the worker 
                    // will shut it down.
                    close(toComponent)
                    done = true
                }
            }
        }
    }
}

// shutdownComponents returns shutdownFunc that shutdowns an 
// array of workers that have been started with the function
// returned by the startComponent function.
func shutdownComponents(out chan<- Value) shutdownFunc {
    return func (shutdownChans [](chan bool)) {
        // Shut down the workers
        for _, schan := range shutdownChans {
            schan <- true
            // The second write is done only after
            // the worker reads the first value, so the worker is
            // not processing a value then.
            schan <- true
        }
        // All workers have now terminated, so I can close the 
        // out channel and let the termination propagate. 
        close(out)
    }
}

// startWriter returns a startFunc that starts a worker that 
// is represented by a writer. 
//
// When the returned function is called it behaves as follows.
// It starts the worker and then forwards the values from the 
// channel in to the worker. When a signal is sent on the end 
// channel the worker is shutdown and the function terminates.
func startWriter(writer Writer, 
                 toWorkers <-chan Value) startFunc {
    return  func(end chan bool) {
        toWriter:= GetChan()
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

// shutdownWriters is a shutdownFunc that shutdowns an 
// array of workers that have been started with the function
// returned by the startWriter function.
func shutdownWriters(shutdownChans [](chan bool)) {
    // Shut down the workers
    for _, schan := range shutdownChans {
        close(schan)
    }
}

// StaticLoadBalancer returns a component that performs static 
// load balancing. That is it starts a nWorkers instances of the
// worker, that can be safely shutdown, when the input channel
// of the returned component is closed.
func StaticLoadBalancer(worker Component, 
                        nWorkers int) Component {
    return func (in <-chan Value, out chan<- Value) {
        toWorkers := GetChan()
        staticLoadBalance(in, toWorkers, 
            startComponent(worker, toWorkers, out), 
            shutdownComponents(out), nWorkers)
    }
}

// StaticLoadBalancer returns a writer that performs static 
// load balancing. That is it starts a nWorkers instances of the
// worker, that can be safely shutdown, when the input channel
// of the returned writer is closed.
func StaticLoadBalancerWriter(worker Writer, 
                              nWorkers int) Writer {
    return func (in <-chan Value) {
        toWorkers := GetChan()
        staticLoadBalance(in, toWorkers, 
            startWriter(worker, toWorkers), 
            shutdownWriters, nWorkers)
    }
}