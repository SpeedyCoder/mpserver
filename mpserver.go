package mpserver

import (
    "log"
    "net/http"
    "os"
    "strings"
)

type Any interface{}

type Value struct {
    Writer http.ResponseWriter
    Request *http.Request
    Result Any
    Done chan<- bool
    ResponseCode int
}

type ValueChan chan Value

//-------------------- Helper Functions ----------------------------
func Listen(s *http.ServeMux, url string, out chan<- Value) {
    s.HandleFunc(url, func (w http.ResponseWriter, r *http.Request) {
        done := make(chan bool)
        w.Header().Set("Server", "mpserver")
        out <- Value{ w, r , nil, done, 0}
        <- done
    })
}

//-------------------- Components ----------------------------------
type Component func (in <-chan Value, out chan<- Value)

// TODO: test this
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

func ConstantComponent(c Any) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            val.Result = c
            out <- val
        }
        close(out)
    }
}

func StringComponent(s string) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            log.Println(val.Request.URL.Path)
            val.Result = s
            out <- val
        }
        close(out)
    }
}

func ErrorComponent(err error) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            val.Result = err
            out <- val
        }
        close(out)
    }    
}

func FileComponent(dir, prefix string) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            // TODO: check if this is safe
            f, err := os.Open(
                dir + strings.TrimPrefix(val.Request.URL.Path, prefix))
            if (err != nil) {
                val.ResponseCode = http.StatusBadRequest
                val.Result = err
                out <- val
                continue
            }

            val.Result = f
            out <- val
        }
        close(out)
    }
}

type Condition func (r *http.Request) bool

func Splitter(cond Condition, in <-chan Value, out1, out2 chan<- Value) {
    for val := range in {
        if (cond(val.Request)) {
            out1 <- val
        } else {
            out2 <- val
        }
    }
    close(out1); close(out2)
}




