package mpserver

import (
    "log"
    "errors"
    "net/http"
    "encoding/json"
)

type Any interface{}

type Value struct {
    Writer http.ResponseWriter
    Request *http.Request
    Result Any
    Done chan bool
}

type ValueChan chan Value

//-------------------- Helper Functions ----------------------------
func Listen(s *http.ServeMux, url string ,out chan Value, done chan bool) {
    s.HandleFunc(url, func (w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Server", "mpserver")
        out <- Value{ w, r , nil, done}
        <- done
    })
}

func ReportError(errChan chan Value, val Value, err error) {
    val.Result = err
    errChan <- val
}

//-------------------- Components ----------------------------------
type Component func (in <-chan Value, out chan<- Value)

func StringComponent(s string) Component {
    return func (in <-chan Value, out chan<- Value) {
        for val := range in {
            val.Result = s
            out <- val
        }
        close(out)
    }
}

//-------------------- Output Writers ------------------------------
func ErrorWriter(in chan Value) {
    for val := range in {
        err, ok := val.Result.(error)
        if (!ok) {
            http.Error(
                val.Writer, 
                "ErrorWriter couldn't write the error", 
                http.StatusInternalServerError)
        } else {
            log.Println(err.Error())
            http.Error(val.Writer, err.Error(), http.StatusInternalServerError)
        }
        val.Done <- true
    }
}

func StringWriter(in chan Value, errChan chan Value) {
    for val := range in {
        s, ok := val.Result.(string)
        if (!ok) {
            ReportError(errChan, val, 
                errors.New("Passed in wrong type to StringWriter."))
        } else {
            val.Writer.Write([]byte(s))
            val.Done <- true
        }   
    }
}

func JsonWriter(in chan Value, errChan chan Value) {
    for val := range in {
        js, err := json.Marshal(val.Result)
        if err != nil {
            ReportError(errChan, val, err)
            return
        }

        (val.Writer).Header().Set("Content-Type", "application/json")
        (val.Writer).Write(js)
        val.Done <- true
    }
}







