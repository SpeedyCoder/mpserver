package mpserver

import (
    "fmt"
	"log"
    "errors"
    "net/http"
    "encoding/json"
    "io"
	"compress/gzip"
    "strings"
)

//-------------------- Helper Functions -------------------------

// writeError writes the error message from the provided error to
// the client that is represented by the provided value.
func writeError(val Value, err error) {
    val.SetResponseCodeIfUndef(http.StatusInternalServerError)
    log.Println("Error:", err.Error())
    http.Error(val.getResponseWriter(), 
        err.Error(), val.getResponseCode())
    val.close()
}

// writeWrongInput is used to write error message back to the 
// client, when a writer is provided with a wrong input type.
func writeWrongInput(val Value, template string) {
    err := errors.New(fmt.Sprintf(template, val.GetResult()))
    writeError(val, err)
}

//-------------------- Output Writers ---------------------------

// Writer is the end of the pipeline which writes results back to
// the clients for all values that it reads from its input 
// channel.
type Writer func (in <-chan Value)

// WriterFunc is used to create custom writers using the 
// MakeWriter function. It should set headers and convert the 
// provided value into a slice of bytes or return an error in 
// case it is provided with a wrong input.
type WriterFunc func (val Value) ([]byte, error)

// MakeWriter generates a Writer using the provided WriterFunc.
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
                writeError(val, err)
            }
        }       
    }
}

// ErrorWriter is a writer used for writing error responses. It 
// expects an error object in the result field of input values. 
func ErrorWriter(in <-chan Value) {
    errorTemplate := "Passed in %t to ErrorWriter."
    for val := range in {
        err, ok := val.GetResult().(error)
        if (!ok) {
            val.SetResponseCode(http.StatusInternalServerError)
            writeWrongInput(val, errorTemplate)
        } else {
            writeError(val, err)
        }
        
    }
}

// StringWriter is a writer used for writing string responses. It 
// expects a string in the result field of input values.  
func StringWriter(in <-chan Value) {
    errorTemplate := "Passed in %t to StringWriter."
    for val := range in {
        s, ok := val.GetResult().(string)
        if (!ok) {
            writeWrongInput(val, errorTemplate)
        } else {
            val.SetResponseCodeIfUndef(http.StatusOK)
        	val.writeHeader()
            val.write([]byte(s))
            val.close()
        }   
    }
}

// JsonWriter is a writer used for writing JSON responses. It 
// expects any object in the result field of input values that 
// can be converted to a JSON string using the json module.
func JsonWriter(in <-chan Value) {
    for val := range in {
        js, err := json.Marshal(val.GetResult())
        if err != nil {
            writeError(val, err)
        } else {
            val.SetHeader("Content-Type", "application/json")
            val.SetResponseCodeIfUndef(http.StatusOK)
            val.writeHeader()
            val.write(js)
            val.close()
        }
    }
}

// GzipWriter is a writer used for writing responses compressed 
// using Gzip compression. It expects an io.ReadCloser object in 
// the result field of input values.
func GzipWriter(in <-chan Value) {
    errorTemplate := "Passed in %t to GzipWriter."
	for val := range in {
		reader, ok := val.GetResult().(io.ReadCloser)
		if (!ok) {
			writeWrongInput(val, errorTemplate)
			continue
		}

        val.SetHeader("Content-Encoding", "gzip")
        val.SetHeader("Content-Type", "application/x-gzip")
        val.SetResponseCodeIfUndef(http.StatusOK)
        val.writeHeader()
		
		gzipWriter := gzip.NewWriter(val.getResponseWriter())
		io.Copy(gzipWriter, reader)
		gzipWriter.Close()
		reader.Close()
		val.close()
	}
}

// GenericWriter is a writer used for writing generic responses.
// It expects an io.ReadCloser object in the result field of 
// input values.
func GenericWriter(in <-chan Value) {
    errorTemplate := "Passed in %t to GenericWriter."
	for val := range in {
		reader, ok := val.GetResult().(io.ReadCloser)
		if (!ok) {
			writeWrongInput(val, errorTemplate)
			continue
		}
		val.SetResponseCodeIfUndef(http.StatusOK)
        val.writeHeader()
		io.Copy(val.getResponseWriter(), reader)
		reader.Close()
		val.close()
	}
}

// ResponseWriter is a writer used for writing generic responses
// that were obtained from another server. It expects a
// mpserver.Response object in the result field of input values.
func ResponseWriter(in <-chan Value) {
    errorTemplate := "Passed in %t to ResponseWriter."
    for val := range in {
        resp, ok := val.GetResult().(Response)
        if (!ok) {
            writeWrongInput(val, errorTemplate)
            continue
        }
        // Set Headers
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

// HttpResponseWriter is a writer used for writing generic 
// responses that were obtained from another server. It expects a
// http.Response object in the result field of input values.
func HttpResponseWriter(in <-chan Value) {
    errorTemplate := "Passed in %t to HttpResponseWriter."
    for val := range in {
        resp, ok := val.GetResult().(*http.Response)
        if (!ok) {
            writeWrongInput(val, errorTemplate)
            continue
        }
        // Set Headers
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