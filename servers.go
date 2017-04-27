package mpserver

import (
	"net/http"
)

func SimpleFileServer(dir, prefix string) http.Handler{
	// Construct the channels
    in := GetChan()
    toFileComp := GetChan()
    out := GetChan()
    toWriter := GetChan()
    errChan := GetChan()

    // Start the file components
    go PathMaker(dir, prefix)(in, toFileComp)
    go FileComponent(toFileComp, out)
    
    // Start the splitter and all the writers
    go ErrorSplitter(out, toWriter, errChan)
    go GenericWriter(toWriter, errChan)
    go ErrorWriter(errChan)

	return Handler(in)
}