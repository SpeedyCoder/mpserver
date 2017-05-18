package mpserver
import "net/http"
import "time"

// FileServerWriter returns a writer, that represents a simple 
// file server that serves files from the directory dir and  
// strips the provided prefix from the requested file paths. 
// The returned writer uses other components internally.
func FileServerWriter(dir, prefix string) Writer {
    return func (in <-chan Job) {
        // Construct the channels
        toFileComp := GetChan()
        out := GetChan()
        toWriter := GetChan()
        errChan := GetChan()

        // Start the file components
        go PathMaker(dir, prefix)(in, toFileComp)
        go FileComponent(toFileComp, out)
        
        // Start the splitter and all the writers
        go ErrorRouter(out, toWriter, errChan)
        go GenericWriter(toWriter)
        ErrorWriter(errChan)
    }
}

// DBalancedFileServer creates a file server and returns an 
// http.Handler object that feeds the incoming requests to this 
// server.
func DBalancedFileServer(dir, prefix string, 
                        maxWorkers int) http.Handler{
    in := GetChan()
    writer := FileServerWriter(dir, prefix)
    lb := DynamicLoadBalancerWriter(
        writer, maxWorkers, time.Nanosecond, time.Minute)

    go lb(in)
	return Handler(in)
}

// SBalancedFileServer creates a file server and returns an 
// http.Handler object that feeds the incoming requests to this 
// server.
func SBalancedFileServer(dir, prefix string, 
                        maxWorkers int) http.Handler{
    in := GetChan()
    writer := FileServerWriter(dir, prefix)
    lb := StaticLoadBalancerWriter(writer, maxWorkers)

    go lb(in)
    return Handler(in)
}

// SBalancedFileServer creates a file server and returns an 
// http.Handler object that feeds the incoming requests to this 
// server.
func LBalancedFileServer(dir, prefix string) http.Handler{
    ins := make([]chan Job, 4)
    writer := FileServerWriter(dir, prefix)
    for i := 0; i < 4; i++ {
        ins[i] = GetChan()
        go writer(ins[i])
    }

    return HandlerTo4(ToOutChans(ins))
}

// SimpleFileServer creates a file server and returns an 
// http.Handler object that feeds the incoming requests to this 
// server.
func SimpleFileServer(dir, prefix string) http.Handler{
    in := GetChan()
    writer := FileServerWriter(dir, prefix)

    go writer(in)
    return Handler(in)
}