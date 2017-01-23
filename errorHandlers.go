package mpserver

import (
    "log"
    "net/http"
    "errors"
)

func ErrorPasser(component Component) Component {
    return func (in <-chan Value, out chan<- Value) {
        toComponent := make(ValueChan)
        go component(toComponent, out)

        for val := range in {
            if _, isErr := val.Result.(error); isErr {
                out <- val
            } else {
                toComponent <- val
            }
        }
        close(toComponent)
        // component will close the out channel
    }    
}

// This can still panic if provided component closes one of its channels
func PannicHandlingComponent(component Component) Component {
    var phComp Component
    phComp = func (in <-chan Value, out chan<- Value) {
        toComponent := make(ValueChan)
        fromComponent := make(ValueChan)
        shutDown := make(chan bool)

        // Recover from panic and restart in a new goroutine
        defer func () {
            if r := recover(); r != nil {
                // Shut down the copying routine
                close(shutDown); close(fromComponent)
                // Start again
                log.Println("Recovered from:", r)
                go phComp(in, out)
            }
        }()
        
        go func () {
            done := false
            var val Value
            var inOpen bool
            for !done {
                select {
                    case val, inOpen = <- in: {
                        if !inOpen {
                            done = true; continue
                        }
                    }
                    case <- shutDown: {
                        done = true; continue
                    }
                }
                toComponent <- val
                if res, ok  := <- fromComponent; ok {
                    out <- res
                } else {
                    val.Result = errors.New("Component crashed.")
                    val.ResponseCode = http.StatusInternalServerError
                    out <- val
                }
                
            }
            // This will cause the component to terminate and then 
            // the out channel will be closed if termination wasn't 
            // caused by a shutdown
            close(toComponent)
        }()

        component(toComponent, fromComponent)
        close(out)
        log.Println("Closed out")
    }

    return phComp
}