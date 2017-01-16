package mpserver

import (
	"log"
    "errors"
    "net/http"
    "encoding/json"
    "io"
	"compress/gzip"
)

//-------------------- Helper Functions ----------------------------

func ReportError(errChan chan<- Value, val Value, err error) {
    val.Result = err
    errChan <- val
}

func GetResponseCode(val Value, defaultCode int) int {
	if (val.ResponseCode == 0) {
		return defaultCode
	}
	return val.ResponseCode
}

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
        	responseCode := GetResponseCode(val, http.StatusInternalServerError)
            log.Println(err.Error())
            http.Error(val.Writer, err.Error(), responseCode)
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
        	responseCode := GetResponseCode(val, http.StatusOK)
        	val.Writer.WriteHeader(responseCode)
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

        responseCode := GetResponseCode(val, http.StatusOK)
        val.Writer.WriteHeader(responseCode)
        (val.Writer).Header().Set("Content-Type", "application/json")
        (val.Writer).Write(js)
        val.Done <- true
    }
}

// TODO: test this
func GzipWriter(in <-chan Value, errChan chan<- Value) {
	for val := range in {
		reader, ok := val.Result.(io.ReadCloser)
		if (!ok) {
			ReportError(errChan, val, 
                errors.New("Passed in wrong type to GzipWriter."))
			continue
		}
		responseCode := GetResponseCode(val, http.StatusOK)
        val.Writer.WriteHeader(responseCode)

		val.Writer.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(val.Writer)
		// Do the compression in a separate goroutine, so that the
		// writer can process another value
		go func() {
			io.Copy(gzipWriter, reader)
			gzipWriter.Close()
			reader.Close()
			val.Done <- true
		} ()
	}
}

