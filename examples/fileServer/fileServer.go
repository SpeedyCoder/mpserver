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

    go mpserver.FileComponent("examples/fileServer", "")(in, toSplitter)
    go mpserver.ErrorSplitter(toSplitter, out, errChan)
    
    go mpserver.Splitter(condition, out, compressed, uncompressed)
    go mpserver.GzipWriter(compressed, errChan)
    go mpserver.GenericWriter(uncompressed, errChan)
    go mpserver.ErrorWriter(errChan)

    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}