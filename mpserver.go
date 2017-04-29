package mpserver

import (
    "net/http"
    "golang.org/x/net/websocket"
    "os"
    "strings"
    "errors"
)

type Any interface{}

type Value struct {
    Request *http.Request
    Writer http.ResponseWriter
    WebSocket *websocket.Conn

    Result Any
    Done chan<- bool
    ResponseCode int
}

type ValueChan chan Value
type InChan <-chan Value
type OutChan chan<- Value

func GetChan() ValueChan {
    return make(ValueChan)
}

//-------------------- HTPP Handlers ----------------------------
func HandlerFunction(out chan<- Value) (func (http.ResponseWriter, *http.Request)) {
    return func (w http.ResponseWriter, r *http.Request) {
        done := make(chan bool)
        w.Header().Set("Server", "mpserver")
        out <- Value{ r, w, nil, nil, done, 200}
        <- done
    }  
}

func Handler(out chan<- Value) http.Handler {
    return http.HandlerFunc(HandlerFunction(out))
}

func Listen(s *http.ServeMux, url string, out chan<- Value) {
    s.HandleFunc(url, HandlerFunction(out))
}

func ListenWebSocket(s *http.ServeMux, url string, out chan<- Value) {
    s.Handle(url, websocket.Handler(func (ws *websocket.Conn) {
        done := make(chan bool)
        out <- Value{ nil, nil, ws, nil, done, 200}
        <- done
    }))
}

//-------------------- Components ----------------------------------
type Component func (in <-chan Value, out chan<- Value)

type ComponentFunc func (val Value) Value

func LinkComponents(components ...Component) Component {
    return func (in <-chan Value, out chan<- Value) {
        iters := len(components) - 1
        if (iters == 0) {
            components[0](in, out)
        } else {
            current := make(ValueChan)
            go components[0](in, current)
            for i:=1; i<iters; i++ {
                next := make(ValueChan)
                go components[i](current, next)
                current = next
            }
            components[iters](current, out)
        }
        
    }
}

func MakeComponent(f ComponentFunc) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            out <- f(val)
        }
        close(out)
    }
}

func ConstantComponent(c Any) Component {
    return MakeComponent(func (val Value) Value {
        val.Result = c; return val
    })
}

func PathMaker(dir, prefix string) Component {
    return MakeComponent(func (val Value) Value {
        val.Result = dir + strings.TrimPrefix(
            val.Request.URL.Path, prefix)
        return val
    })
}

func FileComponent (in <-chan Value, out chan<- Value) {
    handleError := func (val Value, err error) {
        val.ResponseCode = http.StatusBadRequest
        val.Result = err
        out <- val
    }

    for val := range in {
        path, ok := val.Result.(string)
        if (!ok) {
            val.Result = errors.New("No path provided to FileComponent.")
            out <- val
            continue
        }

        // Check if a file exists
        if _, err := os.Stat(path); os.IsNotExist(err) {
            handleError(val, err)
            continue
        }

        // Try to open th file
        file, err := os.Open(path)
        if (err != nil) {
            handleError(val, err)
            continue
        }

        val.Result = file
        out <- val        
    }
    close(out)
}

type Condition func (val Value) bool

func ToOutChans(chans []ValueChan) []chan<- Value {
    res := make([]chan<- Value, len(chans))
    for i, ch := range chans {
        res[i] = ch
    }
    return res
}

func Splitter(in <-chan Value, defOut chan<- Value, outs []chan<- Value, conds []Condition) {
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




