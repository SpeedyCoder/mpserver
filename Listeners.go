package mpserver

import "net/http"

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
        out <- &valueStruct{
            r, nil, UndefinedRespCode, w, nil, done}
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