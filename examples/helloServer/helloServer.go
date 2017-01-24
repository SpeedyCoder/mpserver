package main

import(
    "log"
    "net/http"
    "mpserver"
)

func main() {
    in := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)

    go mpserver.ConstantComponent("Hello world!")(in, out)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)

    mux := http.NewServeMux()
    mpserver.Listen(mux, "/hello", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}