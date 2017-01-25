package main

import(
    "log"
    "net/http"
    "mpserver"
    "strings"
)

func condition(val mpserver.Value) bool {
    return strings.HasSuffix(val.Request.URL.Path, ".go")
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
    
    outs := mpserver.ToOutChans([]mpserver.ValueChan{compressed})
    go mpserver.Splitter(out, uncompressed, outs, []mpserver.Condition{condition})
    go mpserver.GzipWriter(compressed, errChan)
    go mpserver.GenericWriter(uncompressed, errChan)
    go mpserver.ErrorWriter(errChan)

    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}