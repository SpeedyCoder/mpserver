package main

import(
    "log"
    "net/http"
    "mpserver"
)

func main() {
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)
    done := make(chan bool)
    go mpserver.StringComponent("Hello world!")([]mpserver.ValueChan{in}, []mpserver.ValueChan{out})
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/hello", in, done)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}