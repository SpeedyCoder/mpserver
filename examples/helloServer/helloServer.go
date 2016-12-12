package main

import(
    "log"
    "net/http"
    "mpserver"
)

func main() {
    mux := http.NewServeMux()
    in := make(chan mpserver.Value)
    out := make(chan mpserver.Value)
    errChan := make(chan mpserver.Value)
    done := make(chan bool)
    go mpserver.StringComponent(in, out, "Hello world!")
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/hello", in, done)
    log.Println("Listening...")
    http.ListenAndServe(":3000", mux)
}