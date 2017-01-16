package main

import(
    "log"
    "net/http"
    "mpserver"
)

func main() {
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    toSplitter := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)

    fComp := mpserver.FileComponent("examples/fileServer", "")
    go fComp(in, toSplitter)
    go mpserver.ErrorSplitter(toSplitter, out, errChan)
    go mpserver.GzipWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)

    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}