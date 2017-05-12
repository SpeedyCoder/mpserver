package main
import(
    "log"
    "net/http"
    "mpserver"
    "strings"
)
// Test if the provided value contains an error in the Result
func isError(val mpserver.Value) bool {
    _, isErr := val.GetResult().(error)
    return isErr
}
// Test if the path of the requested file ends with .go
func isGoFile(val mpserver.Value) bool {
    return strings.HasSuffix(val.GetRequest().URL.Path, ".go")
}

func main() {
    // Construct the channels
    in := mpserver.GetChan()
    toFileComp := mpserver.GetChan()
    toSplitter := mpserver.GetChan()
    compressed := mpserver.GetChan()
    errChan := mpserver.GetChan()
    splitterOut := mpserver.ToOutChans(
        [](chan mpserver.Value){errChan, compressed})
    uncompressed := mpserver.GetChan()

    // Start the file components
    go mpserver.PathMaker("files", "")(in, toFileComp)
    go mpserver.FileComponent(toFileComp, toSplitter)
    
    // Start the splitter
    go mpserver.Splitter(toSplitter, uncompressed, splitterOut, 
                         []mpserver.Condition{isError, isGoFile})

    // Start the writers
    go mpserver.GzipWriter(compressed)
    go mpserver.GenericWriter(uncompressed)
    go mpserver.ErrorWriter(errChan)

    // Start the server
    mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}