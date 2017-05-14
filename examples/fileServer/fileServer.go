package main
import(
    "mpserver"
    "strings"
)

func main() {
    // Construct the channels
    in := mpserver.GetChan()
    toFileComp := mpserver.GetChan()
    toRouter := mpserver.GetChan()
    compressed := mpserver.GetChan()
    errChan := mpserver.GetChan()
    routerOut := mpserver.ToOutChans(
        [](chan mpserver.Value){errChan, compressed})
    uncompressed := mpserver.GetChan()

    mpserver.Listen("/", in, nil)   // Listener

    // Start the file components
    go mpserver.PathMaker("files", "")(in, toFileComp)
    go mpserver.FileComponent(toFileComp, toRouter)
    
    // Start the router
    go mpserver.Router(toRouter, uncompressed, routerOut, 
                       []mpserver.Condition{isError, isGoFile})

    // Start the writers
    go mpserver.ErrorWriter(errChan)
    go mpserver.GzipWriter(compressed)
    go mpserver.GenericWriter(uncompressed) // File Writer

    mpserver.ListenAndServe(":3000", nil) // Start the server
}

// Condition to test whether the provided value contains an error 
// in the result field
func isError(val mpserver.Value) bool {
    _, isErr := val.GetResult().(error)
    return isErr
}
// Condition to test whether the path of the requested file ends 
// with .go
func isGoFile(val mpserver.Value) bool {
    return strings.HasSuffix(val.GetRequest().URL.Path, ".go")
}