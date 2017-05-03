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
func WriteError(val Value, err error) {
    val.SetResponseCodeIfUndef(http.StatusInternalServerError)
    log.Println("Error:", err.Error())
    http.Error(val.getResponseWriter(), err.Error(), val.getResponseCode())
    val.close()
}

//-------------------- Output Writers ------------------------------
type Writer func (in <-chan Value)

// This should set response code if it returns true
type WriterFunc func (val Value) ([]byte, error)

func MakeWriter(writerFunc WriterFunc) Writer {
    return func (in <-chan Value) {
        for val := range in {
            resp, err := writerFunc(val)

            if (err == nil) {
                // Write the response
                val.SetResponseCodeIfUndef(http.StatusOK)
                val.writeHeader()
                val.write(resp)
                val.close()
            } else {
                WriteError(val, err)
            }
        }       
    }
}

func AddErrorSplitter(writer Writer, errChan chan<- Value) Writer {
    return func (in <-chan Value) {
        toWriter := make(ValueChan)
        go ErrorSplitter(in, toWriter, errChan)
        writer(toWriter)
    }
}

func ErrorWriter(in <-chan Value) {
    for val := range in {
        err, ok := val.GetResult().(error)
        if (!ok) {
            http.Error(
                val.getResponseWriter(), 
                "ErrorWriter couldn't write the error", 
                http.StatusInternalServerError)
        } else {
            val.SetResponseCodeIfUndef(http.StatusInternalServerError)
            log.Println(err.Error())
            http.Error(val.getResponseWriter(), err.Error(), val.getResponseCode())
        }
        val.close()
    }
}

func ErrorSplitter(in <-chan Value, out chan<- Value, errChan chan<- Value) {
    for val := range in {
        if _, ok := val.GetResult().(error); ok {
            errChan <- val
        } else {
            out <- val
        }
    }
    close(out)
}

func StringWriter(in <-chan Value) {
    for val := range in {
        s, ok := val.GetResult().(string)
        if (!ok) {
            WriteError(val, 
                errors.New("Passed in wrong type to StringWriter."))
        } else {
            val.SetResponseCodeIfUndef(http.StatusOK)
        	val.writeHeader()
            val.write([]byte(s))
            val.close()
        }   
    }
}

func JsonWriter(in <-chan Value) {
    for val := range in {
        js, err := json.Marshal(val.GetResult())
        if err != nil {
            WriteError(val, err)
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
func GzipWriter(in <-chan Value) {
	for val := range in {
		reader, ok := val.GetResult().(io.ReadCloser)
		if (!ok) {
			WriteError(val, 
                errors.New("Passed in wrong type to GzipWriter."))
			continue
		}

        val.SetHeader("Content-Encoding", "gzip")
        val.SetHeader("Content-Type", "application/x-gzip")
        val.SetResponseCodeIfUndef(http.StatusOK)
        val.writeHeader()
		
		gzipWriter := gzip.NewWriter(val.getResponseWriter())
		// Maybe do the compression in a separate goroutine, so that the
		// writer can process another value
		io.Copy(gzipWriter, reader)
		gzipWriter.Close()
		reader.Close()
		val.close()
	}
}

func GenericWriter(in <-chan Value) {
	for val := range in {
		reader, ok := val.GetResult().(io.ReadCloser)
		if (!ok) {
			WriteError(val, 
                errors.New("Passed in wrong type to Writer."))
			continue
		}
		val.SetResponseCodeIfUndef(http.StatusOK)
        val.writeHeader()

		// Maybe do the writing in a separate goroutine, so that the
		// writer can process another value
		io.Copy(val.getResponseWriter(), reader)
		reader.Close()
		val.close()
	}
}

func ResponseWriter(in <-chan Value) {
    for val := range in {
        resp, ok := val.GetResult().(Response)
        if (!ok) {
            WriteError(val, 
                errors.New("Passed in wrong type to ResponseWriter."))
            continue
        }

        // Write Headers
        header := val.getResponseWriter().Header()
        for key, value := range resp.Header {
            header.Set(key, strings.Join(value, ""))
        }

        val.SetResponseCodeIfUndef(resp.ResponseCode)
        val.writeHeader()
        val.write(resp.Body)
        val.close()
    }
}

func HttpResponseWriter(in <-chan Value) {
    for val := range in {
        resp, ok := val.GetResult().(*http.Response)
        if (!ok) {
            WriteError(val, 
                errors.New("Passed in wrong type to HttpResponseWriter."))
            continue
        }
        // Write Headers
        header := val.getResponseWriter().Header()
        for key, value := range resp.Header {
            header.Set(key, strings.Join(value, ""))
        }
        val.writeHeader()

        func() {
            defer resp.Body.Close()
            io.Copy(val.getResponseWriter(), resp.Body)
            val.close()
        } ()
    }
}


