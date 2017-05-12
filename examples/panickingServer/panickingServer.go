package main

import(
    "log"
    "net/http"
    "mpserver"
)

func panickingComponent(in <-chan mpserver.Value, 
                        out chan<- mpserver.Value) {
    i := 0
    for val := range in {
        val.SetResult("Hello World!")
        if i >= 2 {
            panic("There were more than 2 requests.")
        }
        out <- val
        i++
    }
    close(out)
}

func main() {
    toComponent := mpserver.GetChan()
    toSplitter := mpserver.GetChan()
    toStringWriter := mpserver.GetChan()
    toErrorWriter := mpserver.GetChan()

    phComp := mpserver.PannicHandler(panickingComponent)
    go phComp(toComponent, toSplitter)
    go mpserver.ErrorSplitter(
        toSplitter, toStringWriter, toErrorWriter)
    go mpserver.StringWriter(toStringWriter)
    go mpserver.ErrorWriter(toErrorWriter)

    mux := http.NewServeMux()
    mpserver.Listen(mux, "/", toComponent)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}