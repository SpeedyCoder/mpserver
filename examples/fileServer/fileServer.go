package main

import(
    "log"
    "net/http"
    "mpserver"
    "strings"
)

func condition(r *http.Request) bool {
    return strings.HasSuffix(r.URL.Path, ".go")
}

func main() {
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    toSplitter := make(mpserver.ValueChan)
    compressed := make(mpserver.ValueChan)
    uncompressed := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)

    fComp := mpserver.FileComponent("examples/fileServer", "")
    go fComp(in, toSplitter)
    go mpserver.Splitter(condition, toSplitter, compressed, uncompressed)
    go mpserver.ErrorSplitter(compressed, out, errChan)
    go mpserver.GzipWriter(out, errChan)
    go mpserver.Writer(uncompressed, errChan)
    go mpserver.ErrorWriter(errChan)

    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}