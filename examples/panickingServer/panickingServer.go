package main

import(
    "log"
    "net/http"
    "mpserver"
)

func panickingComponent(in <-chan mpserver.Value, out chan<- mpserver.Value) {
    i := 0
    for val := range in {
        val.Result = "Hello World!"
        if i >= 2 {
            panic("There were more than 2 requests.")
        }
        out <- val
        i++
    }
}

func main() {
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    toSplitter := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)
    phComp := mpserver.PannicHandlingComponent(panickingComponent)
    go phComp(in, toSplitter)
    go mpserver.ErrorSplitter(toSplitter, out, errChan)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}