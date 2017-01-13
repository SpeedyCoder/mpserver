package main

import(
    "log"
    "net/http"
    "mpserver"
    // "time"
)

func main() {
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)
    done := make(chan bool)
    sComp := mpserver.StringComponent("Hello world!")
    // lbComp := mpserver.LoadBalancingComponent(time.Second, time.Second*5, sComp)
    cComp := mpserver.CacheComponent(sComp)
    go cComp(in, out)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/hello", in, done)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}