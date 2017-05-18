package mpserver

import "net/http"
import "log"

const UndefinedRespCode int = -1;

// DefaultServeMux is used by Listen and ListenAndServe
// functions, when a ServeMux or handler object isn't specified.
var DefaultServeMux = http.NewServeMux()

// GetChan returns an unbuffered channel of type chan Job.
func GetChan() chan Job {
    return make(chan Job)
}

// ----------------------- HTPP Handlers ------------------------

// HandlerFunction takes an output channel and returns a function
// that can be used with the http.HandlerFunc to generate a
// http.Handler. This handler will for each incoming request 
// create a Job object and send it to the output channel.
func HandlerFunction(out chan<- Job) (
    func (http.ResponseWriter, *http.Request)) {
    return func (w http.ResponseWriter, r *http.Request) {
        done := make(chan bool)
        w.Header().Set("Server", "mpserver")
        out <- &jobStruct{
            r, nil, UndefinedRespCode, w, nil, done}
        <- done
    }  
}

// Handler returns an http.Handler object that for each incoming 
// request creates a Job object and sends it to the output 
// channel.
func Handler(out chan<- Job) http.Handler {
    return http.HandlerFunc(HandlerFunction(out))
}

// Handler returns an http.Handler object that for each incoming 
// request creates a Job object and sends it to one of the output 
// channels.
func HandlerTo4(outs []chan<- Job) http.Handler {
    if (len(outs) != 4) {
        panic("Incorrect number of channels.")
    }
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        done := make(chan bool)
        w.Header().Set("Server", "mpserver")
        job := &jobStruct{
            r, nil, UndefinedRespCode, w, nil, done}

        select {
            case outs[0] <- job: {}
            case outs[1] <- job: {}
            case outs[2] <- job: {}
            case outs[3] <- job: {}
        }
        <- done
    })
}

// Listen registers a handler on the provided ServeMux for the 
// provided url, that will for each incoming request create 
// a Job object and send it to the output channel. If the 
// provided ServeMux is nil DefaultServeMux is used.
func Listen(url string, out chan<- Job, mux *http.ServeMux) {
    if (mux != nil) {
        mux.HandleFunc(url, HandlerFunction(out))
    } else {
        DefaultServeMux.HandleFunc(url, HandlerFunction(out))
    }
}

// ListenAndServe calls the ListenAndServe method of the default
// http package with the provided handler if the handler is not 
// nil. It otherwise uses the DefaultServeMux.
func ListenAndServe(addr string, handler http.Handler) error {
    log.Println("Listening at", addr)
    if (handler != nil) {
        return http.ListenAndServe(addr, handler)
    }
    return http.ListenAndServe(addr, DefaultServeMux)
}
