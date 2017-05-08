package mpserver

import (
	"net/http"
    "time"
)

func FileServerWriter(dir, prefix string) Writer {
    return func (in <-chan Value) {
        // Construct the channels
        toFileComp := GetChan()
        out := GetChan()
        toWriter := GetChan()
        errChan := GetChan()

        // Start the file components
        go PathMaker(dir, prefix)(in, toFileComp)
        go FileComponent(toFileComp, out)
        
        // Start the splitter and all the writers
        go ErrorSplitter(out, toWriter, errChan)
        go GenericWriter(toWriter)
        go ErrorWriter(errChan)
    }
}

func SimpleFileServer(dir, prefix string, maxWorkers int) http.Handler{
    in := GetChan()
    writer := FileServerWriter(dir, prefix)
    lb := DynamicLoadBalancerW(
        time.Microsecond, time.Second*10, writer, maxWorkers)

    go lb(in)
	return Handler(in)
}


