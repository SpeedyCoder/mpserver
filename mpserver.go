package main

import (
  	"log"
    "errors"
  	"net/http"
    "encoding/json"
)

type Any interface{}

type Value struct {
    writer http.ResponseWriter
    request *http.Request
    value Any
    done chan bool
}

//-------------------- Helper Functions ----------------------------
func Link(s *http.ServeMux, url string ,out chan Value, done chan bool) {
    s.HandleFunc(url, func (w http.ResponseWriter, r *http.Request) {
        // w.Header().Set("Server", "mpserver")
        out <- Value{ w, r , nil, done}
        <- done
    })
}

func ReportError(errChan chan Value, val Value, err error) {
    val.value = err
    errChan <- val
}

//-------------------- Components ----------------------------------

//-------------------- Output Writers ------------------------------
func ErrorWriter(in chan Value) {
    for val := range in {
        err, ok := val.value.(error)
        if (!ok) {
            http.Error(
                val.writer, 
                "ErrorWriter couldn't write the error", 
                http.StatusInternalServerError)
        } else {
            http.Error(val.writer, err.Error(), http.StatusInternalServerError)
        }
        val.done <- true
    }
}

func StringWriter(in chan Value, errChan chan Value) {
    for val := range in {
        s, ok := val.value.(string)
        if (!ok) {
            ReportError(errChan, val, 
                errors.New("Passed in wrong type to StringWriter."))
        } else {
            log.Println(s)
            val.writer.Write([]byte(s))
            log.Println("Written: " + s)
            val.done <- true
        }   
    }
}

func JsonWriter(in chan Value, errChan chan Value) {
    for val := range in {
        js, err := json.Marshal(val.value)
        if err != nil {
            ReportError(errChan, val, err)
            return
        }

        (val.writer).Header().Set("Content-Type", "application/json")
        (val.writer).Write(js)
        val.done <- true
    }
}

func Hello(in chan Value, out chan Value) {
    for val := range in {
        val.value = "Hello world!"
        out <- val
    }
    close(out)
}

func main() {
    mux := http.NewServeMux()
    in := make(chan Value)
    out := make(chan Value)
    errChan := make(chan Value)
    done := make(chan bool)
    go Hello(in, out)
    go StringWriter(out, errChan)
    go ErrorWriter(errChan)
    Link(mux, "/hello", in, done)
    log.Println("Listening...")
    http.ListenAndServe(":3000", mux)
}









