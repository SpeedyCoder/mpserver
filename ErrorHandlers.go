package mpserver

import (
    "log"
    "net/http"
    "errors"
)

// ErrorPasser returns a component that forwards to the worker 
// only the jobs that don't contain an error object in the 
// result field. Jobs with error result are just passed further 
// down the network.
func ErrorPasser(worker Component) Component {
    return func (in <-chan Job, out chan<- Job) {
        toComponent := GetChan()
        go worker(toComponent, out)

        for job := range in {
            if _, isErr := job.GetResult().(error); isErr {
                out <- job
            } else {
                toComponent <- job
            }
        }
        // Close the channel to the worker and the worker 
        // will close the output channel.
        close(toComponent)
    }    
}

// PannicHandler takes a worker component, which can cause panic.
// If that happens the component writes an error to the result 
// field of the job that caused the panic, outputs it and then 
// restarts the worker.
func PannicHandler(worker Component) Component {
    var phComp Component
    phComp = func (in <-chan Job, out chan<- Job) {
        toComponent := GetChan()
        fromComponent := GetChan()
        shutDown := make(chan bool)

        // Recover from panic and restart in a new goroutine.
        defer func () {
            if r := recover(); r != nil {
                // Shut down the copying goroutine
                close(shutDown); close(fromComponent)
                // Start again
                log.Println("Recovered from:", r)
                go phComp(in, out)
            }
        }()
        
        // Goroutine that forwards incoming jobs to worker and
        // then collects the result and sends it on the output 
        // channel. If the component panics while processing a 
        // job this goroutine sends an error on the output 
        // channel.
        go func () {
            done := false
            var job Job
            var inOpen bool
            for !done {
                select {
                    case job, inOpen = <- in: {
                        if !inOpen {
                            // Normal shut down.
                            done = true; continue
                        }
                    }
                    case <- shutDown: {
                        // The component crashed when it wasn't 
                        // processing a job.
                        done = true; continue
                    }
                }
                toComponent <- job
                if res, ok  := <- fromComponent; ok {
                    out <- res
                } else {
                    // Worker panicked, so we need to report
                    // an error.
                    done = true
                    job.SetResult(
                        errors.New("Component crashed."))
                    job.SetResponseCode(
                        http.StatusInternalServerError)
                    out <- job
                }
                
            }
            // This will make the component terminate and then 
            // the out channel will be closed if termination 
            // wasn't caused by a shutdown.
            close(toComponent)
        }()

        worker(toComponent, fromComponent)
        close(out)
    }

    return phComp
}