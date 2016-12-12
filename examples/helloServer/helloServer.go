package main

import(
    "log"
    "net/http"
    "mpserver"
)

func Hello(in chan mpserver.Value, out chan mpserver.Value) {
    for val := range in {
        val.Result = "Hello world!"
        out <- val
    }
    close(out)
}

func main() {
    mux := http.NewServeMux()
    in := make(chan mpserver.Value)
    out := make(chan mpserver.Value)
    errChan := make(chan mpserver.Value)
    done := make(chan bool)
    go Hello(in, out)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/hello", in, done)
    log.Println("Listening...")
    http.ListenAndServe(":3000", mux)
}