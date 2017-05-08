package mpserver

import (
    "net/http"
    "golang.org/x/net/websocket"
    "os"
    "strings"
    "errors"
    "reflect"
)

const UndefinedRespCode int = -1;

//------------------------ Value definition ---------------------------
type Value interface {
    //Public methods
    GetRequest() *http.Request
    GetResult() interface{}
    SetResult(interface{})
    SetResponseCode(int)
    SetResponseCodeIfUndef(int)
    SetHeader(string, string)

    // Private methods
    getResponseWriter() http.ResponseWriter
    getResponseCode() int
    writeHeader()
    write([]byte)
    close()
}

type valueStruct struct {
    request *http.Request
    result interface{}

    responseCode int
    responseWriter http.ResponseWriter
    webSocket *websocket.Conn
    done chan<- bool
}

func (val valueStruct) GetRequest() *http.Request {
    return val.request
}

func (val valueStruct) GetResult() interface{} {
    return val.result
}

func (val *valueStruct) SetResult(newResult interface{}) {
    val.result = newResult
}

func (val *valueStruct) SetResponseCode(responseCode int) {
    val.responseCode = responseCode;
}

func (val *valueStruct) SetResponseCodeIfUndef(responseCode int) {
    if (val.responseCode == UndefinedRespCode) {
        val.responseCode = responseCode;
    }
}

func (val valueStruct) SetHeader(key, value string) {
    val.responseWriter.Header().Set(key, value)
}

func (val valueStruct) getResponseWriter() http.ResponseWriter {
    return val.responseWriter
}

func (val valueStruct) getResponseCode() int {
    return val.responseCode
}

func (val valueStruct) writeHeader() {
    val.responseWriter.WriteHeader(val.responseCode);
}

func (val valueStruct) write(body []byte) {
    val.responseWriter.Write(body)
}

func (val valueStruct) close() {
    val.done <- true;
}

type ValueChan chan Value
type InChan <-chan Value
type OutChan chan<- Value

func GetChan() ValueChan {
    return make(ValueChan)
}

//------------------------ HTPP Handlers ------------------------------
func HandlerFunction(out chan<- Value) (func (http.ResponseWriter, *http.Request)) {
    return func (w http.ResponseWriter, r *http.Request) {
        done := make(chan bool)
        w.Header().Set("Server", "mpserver")
        out <- &valueStruct{ r, nil, UndefinedRespCode, w, nil, done}
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
        out <- &valueStruct{ nil, nil, UndefinedRespCode, nil, ws, done}
        <- done
    }))
}

//----------------------- Components ----------------------------------
type Component func (in <-chan Value, out chan<- Value)

type ComponentFunc func (val Value)

func MakeComponent(f ComponentFunc) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            f(val); out <- val
        }
        close(out)
    }
}

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

func ConstantComponent(c interface{}) Component {
    return MakeComponent(func (val Value) {
        val.SetResult(c)
    })
}

func PathMaker(dir, prefix string) Component {
    return MakeComponent(func (val Value) {
        val.SetResult(dir + strings.TrimPrefix(
            val.GetRequest().URL.Path, prefix))
    })
}

func FileComponent (in <-chan Value, out chan<- Value) {
    handleError := func (val Value, err error) {
        val.SetResponseCode(http.StatusBadRequest)
        val.SetResult(err)
        out <- val
    }

    for val := range in {
        path, ok := val.GetResult().(string)
        if (!ok) {
            val.SetResult(errors.New("No path provided to FileComponent."))
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

        val.SetResult(file)
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

// Reads values from multiple sources and sends them to a single output
func Collector(ins []<-chan Value, out chan<- Value) {
    cases := make([]reflect.SelectCase, len(ins))
    for i := 0; i < len(ins); i++ {
        cases[i] = reflect.SelectCase{
            Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ins[i])}
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




