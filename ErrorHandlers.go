package mpserver

import (
    "log"
    "net/http"
    "errors"
)

// ErrorPasser returns a component that forwards to the worker 
// only the values that don't contain an error object in the 
// result field. Error values are just passed further down the 
// network.
func ErrorPasser(worker Component) Component {
    return func (in <-chan Value, out chan<- Value) {
        toComponent := GetChan()
        go worker(toComponent, out)

        for val := range in {
            if _, isErr := val.GetResult().(error); isErr {
                out <- val
            } else {
                toComponent <- val
            }
        }
        // Close the channel to the worker and the worker 
        // will close the output channel.
        close(toComponent)
    }    
}

// PannicHandler takes a worker component, which can cause panic.
// If that happens the component writes an error to the value 
// that caused the panic, outputs it and then restarts the 
// worker.
func PannicHandler(worker Component) Component {
    var phComp Component
    phComp = func (in <-chan Value, out chan<- Value) {
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
        
        // Goroutine that forwards incoming values to worker and
        // then collects the result and sends it on the output 
        // channel. If the component panics while processing a 
        // value this goroutine sends an error on the output 
        // channel.
        go func () {
            done := false
            var val Value
            var inOpen bool
            for !done {
                select {
                    case val, inOpen = <- in: {
                        if !inOpen {
                            // Normal shut down.
                            done = true; continue
                        }
                    }
                    case <- shutDown: {
                        // Then component crashed when it wasn't 
                        // processing a value.
                        done = true; continue
                    }
                }
                toComponent <- val
                if res, ok  := <- fromComponent; ok {
                    out <- res
                } else {
                    // Worker panicked, so we need to report
                    // an error.
                    done = true
                    val.SetResult(
                        errors.New("Component crashed."))
                    val.SetResponseCode(
                        http.StatusInternalServerError)
                    out <- val
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