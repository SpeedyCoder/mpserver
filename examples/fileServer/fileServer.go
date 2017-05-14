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
        [](chan mpserver.Job){errChan, compressed})
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

// Condition to test whether the provided job contains an error 
// in the result field
func isError(job mpserver.Job) bool {
    _, isErr := job.GetResult().(error)
    return isErr
}
// Condition to test whether the path of the requested file ends 
// with .go
func isGoFile(job mpserver.Job) bool {
    return strings.HasSuffix(job.GetRequest().URL.Path, ".go")
}