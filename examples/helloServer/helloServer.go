package main

import(
    "log"
    "net/http"
    "mpserver"
)

func main() {
    in := mpserver.GetChan()
    out := mpserver.GetChan()

    mux := http.NewServeMux()
    mpserver.Listen(mux, "/hello", in)

    go mpserver.ConstantComponent("Hello world!")(in, out)
    go mpserver.StringWriter(out, nil)
    
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}