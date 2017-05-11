package mpserver

import (
    "net/http"
    "reflect"
)

const UndefinedRespCode int = -1;

// GetChan returns an unbuffered channel of type chan Value.
func GetChan() chan Value {
    return make(chan Value)
}

//------------------------ HTPP Handlers ------------------------

// HandlerFunction takes an output channel and returns a function
// that can be used with the http.HandlerFunc to generate a
// http.Handler. This handler will for each incoming request 
// create a Value object and send it to the output channel.
func HandlerFunction(out chan<- Value) (
    func (http.ResponseWriter, *http.Request)) {
    return func (w http.ResponseWriter, r *http.Request) {
        done := make(chan bool)
        w.Header().Set("Server", "mpserver")
        out <- &valueStruct{ r, nil, UndefinedRespCode, w, nil, done}
        <- done
    }  
}

// Handler returns an http.Handler object that for each incoming 
// request creates a Value object and sends it to the output 
// channel.
func Handler(out chan<- Value) http.Handler {
    return http.HandlerFunc(HandlerFunction(out))
}

// Listen registers a handler on the provided ServeMux for the 
// provided url, that will for each incoming request create 
// a Value object and send it to the output channel.
func Listen(s *http.ServeMux, url string, out chan<- Value) {
    s.HandleFunc(url, HandlerFunction(out))
}

// func ListenWebSocket(s *http.ServeMux, url string, out chan<- Value) {
//     s.Handle(url, websocket.Handler(func (ws *websocket.Conn) {
//         done := make(chan bool)
//         out <- &valueStruct{ nil, nil, UndefinedRespCode, nil, ws, done}
//         <- done
//     }))
// }

type Condition func (val Value) bool

func ToOutChans(chans [](chan Value)) []chan<- Value {
    res := make([]chan<- Value, len(chans))
    for i, ch := range chans {
        res[i] = ch
    }
    return res
}

func Splitter(in <-chan Value, defOut chan<- Value, 
              outs []chan<- Value, conds []Condition) {
    if len(outs) != len(conds) {
        panic("Number of channels and conditions is not equal.")
    }

    for val := range in {
        sent := false
        for i, cond := range conds {
            if cond(val) {
                outs[i] <- val; sent = true
                break
            }
        }

        if !sent {
            defOut <- val
        }
    }

    for _, ch := range outs {
        close(ch)
    }
}

// Reads values from multiple sources and sends them to a single output
func Collector(ins []<-chan Value, out chan<- Value) {
    cases := make([]reflect.SelectCase, len(ins))
    for i := 0; i < len(ins); i++ {
        cases[i] = reflect.SelectCase{
            Dir: reflect.SelectRecv, 
            Chan: reflect.ValueOf(ins[i])}
    }

    done := false
    for !done {
        index, value, ok := reflect.Select(cases)
        if (ok) {
            val := value.Interface().(Value)
            out <- val
        } else {
            cases = append(cases[:index], cases[index+1:]...)
            if len(cases) == 0 {
                done = true
            }
        }
    }
    close(out)
}




