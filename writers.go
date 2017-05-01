package mpserver

import (
	"log"
    "errors"
    "net/http"
    "encoding/json"
    "io"
	"compress/gzip"
    "strings"
)

//-------------------- Helper Functions ----------------------------
func ReportError(errChan chan<- Value, val Value, err error) {
    val.Result = err
    errChan <- val
}

//-------------------- Output Writers ------------------------------
type Writer func (in <-chan Value, errChan chan<- Value)

// This should set response code if it returns true
type WriterFunc func (val *Value) ([]byte, bool)

func MakeWriter(writer WriterFunc) Writer {
    return func (in <-chan Value, errChan chan<- Value) {
        for val := range in {
            resp, doWrite := writer(&val)

            if (doWrite) {
                // Write the response
                val.writeHeader()
                val.write(resp)
                val.close()
            } else {
                // Report Error
                errChan <- val
            }
        }       
    }
}

func AddErrorSplitter(writer Writer) Writer {
    return func (in <-chan Value, errChan chan<- Value) {
        toWriter := make(ValueChan)
        go ErrorSplitter(in, toWriter, errChan)
        writer(toWriter, errChan)
    }
    
}

func ErrorWriter(in <-chan Value) {
    for val := range in {
        err, ok := val.Result.(error)
        if (!ok) {
            http.Error(
                val.responseWriter, 
                "ErrorWriter couldn't write the error", 
                http.StatusInternalServerError)
        } else {
            val.SetResponseCodeIfUndef(http.StatusInternalServerError)
            log.Println(err.Error())
            http.Error(val.responseWriter, err.Error(), val.responseCode)
        }
        val.close()
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
    close(out)
}

func StringWriter(in <-chan Value, errChan chan<- Value) {
    for val := range in {
        s, ok := val.Result.(string)
        if (!ok) {
            log.Printf("%t", val.Result)
            ReportError(errChan, val, 
                errors.New("Passed in wrong type to StringWriter."))
        } else {
            val.SetResponseCodeIfUndef(http.StatusOK)
        	val.writeHeader()
            val.write([]byte(s))
            val.close()
        }   
    }
}

func JsonWriter(in <-chan Value, errChan chan<- Value) {
    for val := range in {
        js, err := json.Marshal(val.Result)
        if err != nil {
            ReportError(errChan, val, err)
        } else {
            val.SetHeader("Content-Type", "application/json")
            val.SetResponseCodeIfUndef(http.StatusOK)
            val.writeHeader()
            val.write(js)
            val.close()
        }
    }
}

// TODO: sort out Content-Type header
func GzipWriter(in <-chan Value, errChan chan<- Value) {
	for val := range in {
		reader, ok := val.Result.(io.ReadCloser)
		if (!ok) {
			ReportError(errChan, val, 
                errors.New("Passed in wrong type to GzipWriter."))
			continue
		}

        val.SetHeader("Content-Encoding", "gzip")
        val.SetHeader("Content-Type", "application/x-gzip")
        val.SetResponseCodeIfUndef(http.StatusOK)
        val.writeHeader()
		
		gzipWriter := gzip.NewWriter(val.responseWriter)
		// Maybe do the compression in a separate goroutine, so that the
		// writer can process another value
		io.Copy(gzipWriter, reader)
		gzipWriter.Close()
		reader.Close()
		val.close()
	}
}

func GenericWriter(in <-chan Value, errChan chan<- Value) {
	for val := range in {
		reader, ok := val.Result.(io.ReadCloser)
		if (!ok) {
			ReportError(errChan, val, 
                errors.New("Passed in wrong type to Writer."))
			continue
		}
		val.SetResponseCodeIfUndef(http.StatusOK)
        val.writeHeader()

		// Maybe do the writing in a separate goroutine, so that the
		// writer can process another value
		io.Copy(val.responseWriter, reader)
		reader.Close()
		val.close()
	}
}

func ResponseWriter(in <-chan Value, errChan chan<- Value) {
    for val := range in {
        resp, ok := val.Result.(Response)
        if (!ok) {
            ReportError(errChan, val, 
                errors.New("Passed in wrong type to ResponseWriter."))
            continue
        }

        // Write Headers
        header := val.responseWriter.Header()
        for key, value := range resp.Header {
            header.Set(key, strings.Join(value, ""))
        }

        val.SetResponseCodeIfUndef(resp.ResponseCode)
        val.writeHeader()
        val.write(resp.Body)
        val.close()
    }
}

func HttpResponseWriter(in <-chan Value, errChan chan<- Value) {
    for val := range in {
        resp, ok := val.Result.(*http.Response)
        if (!ok) {
            ReportError(errChan, val, 
                errors.New("Passed in wrong type to HttpResponseWriter."))
            continue
        }
        // Write Headers
        header := val.responseWriter.Header()
        for key, value := range resp.Header {
            header.Set(key, strings.Join(value, ""))
        }
        val.writeHeader()

        func() {
            defer resp.Body.Close()
            io.Copy(val.responseWriter, resp.Body)
            val.close()
        } ()
    }
}


