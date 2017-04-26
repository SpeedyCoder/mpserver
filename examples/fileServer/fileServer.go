package main

import(
    "log"
    "net/http"
    "mpserver"
    "strings"
)

func isError(val mpserver.Value) bool {
    _, isErr := val.Result.(error)
    return isErr
}

func isGoFile(val mpserver.Value) bool {
    return strings.HasSuffix(val.Request.URL.Path, ".go")
}

func main() {
    // Construct the channels
    in := mpserver.GetChan()
    toFileComp := mpserver.GetChan()
    toSplitter := mpserver.GetChan()
    compressed := mpserver.GetChan()
    errChan := mpserver.GetChan()
    splitterOut := mpserver.ToOutChans([]mpserver.ValueChan{errChan, compressed})
    uncompressed := mpserver.GetChan()

    // Start the file components
    go mpserver.PathMaker("examples/fileServer", "")(in, toFileComp)
    go mpserver.FileComponent(toFileComp, toSplitter)
    
    // Start the splitter and all the writers
    go mpserver.Splitter(toSplitter, uncompressed, splitterOut, 
                         []mpserver.Condition{isError, isGoFile})
    go mpserver.GzipWriter(compressed, errChan)
    go mpserver.GenericWriter(uncompressed, errChan)
    go mpserver.ErrorWriter(errChan)

    // Start the server
    mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}