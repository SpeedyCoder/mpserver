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
    go mpserver.StringWriter(out)
    
    log.Println("Listening on port 8080...")
    http.ListenAndServe(":8080", mux)
}