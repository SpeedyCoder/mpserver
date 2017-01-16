package mpserver

import (
	"log"
    "errors"
    "net/http"
    "encoding/json"
    "io"
	"compress/gzip"
)

//-------------------- Output Writers ------------------------------
func ErrorWriter(in <-chan Value) {
    for val := range in {
        err, ok := val.Result.(error)
        if (!ok) {
            http.Error(
                val.Writer, 
                "ErrorWriter couldn't write the error", 
                http.StatusInternalServerError)
        } else {
            log.Println(err.Error())
            http.Error(val.Writer, err.Error(), http.StatusInternalServerError)
        }
        val.Done <- true
    }
}

func ErrorSplitter(in <-chan Value, out chan<- Value, errChan chan<- Value) {
    for val := range in {
        if _, ok := val.Result.(error); ok {
            errChan <- val
        } else {
            out <- val
        }
    }
    close(errChan)
    close(out)
}

func StringWriter(in <-chan Value, errChan chan<- Value) {
    for val := range in {
        s, ok := val.Result.(string)
        if (!ok) {
            ReportError(errChan, val, 
                errors.New("Passed in wrong type to StringWriter."))
        } else {
            val.Writer.Write([]byte(s))
            val.Done <- true
        }   
    }
}

func JsonWriter(in <-chan Value, errChan chan<- Value) {
    for val := range in {
        js, err := json.Marshal(val.Result)
        if err != nil {
            ReportError(errChan, val, err)
            return
        }

        (val.Writer).Header().Set("Content-Type", "application/json")
        (val.Writer).Write(js)
        val.Done <- true
    }
}

// TODO: test this
func GzipWriter(in <-chan Value, errChan chan<- Value) {
	for val := range in {
		reader, ok := val.Result.(io.Reader)
		if (!ok) {
			ReportError(errChan, val, 
                errors.New("Passed in wrong type to GzipWriter."))
			continue
		}
		val.Writer.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(val.Writer)
		go func() {
			io.Copy(gzipWriter, reader)
			gzipWriter.Close()
			val.Done <- true
		} ()
	}
}

