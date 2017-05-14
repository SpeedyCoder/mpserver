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
// the client that is represented by the provided job.
func writeError(job Job, err error) {
    job.SetResponseCodeIfUndef(http.StatusInternalServerError)
    log.Println("Error:", err.Error())
    http.Error(job.getResponseWriter(), 
        err.Error(), job.getResponseCode())
    job.close()
}

// writeWrongInput is used to write error message back to the 
// client, when a writer is provided with a wrong input type.
func writeWrongInput(job Job, template string) {
    err := errors.New(fmt.Sprintf(template, job.GetResult()))
    writeError(job, err)
}

//-------------------- Output Writers ---------------------------

// Writer is the end of the pipeline which writes results back to
// the clients for all jobs that it reads from its input channel.
type Writer func (in <-chan Job)

// WriterFunc is used to create custom writers using the 
// MakeWriter function. It should set headers and convert the 
// provided job into a slice of bytes or return an error in 
// case it is provided with a wrong input.
type WriterFunc func (job Job) ([]byte, error)

// MakeWriter generates a Writer using the provided WriterFunc.
func MakeWriter(writerFunc WriterFunc) Writer {
    return func (in <-chan Job) {
        for job := range in {
            resp, err := writerFunc(job)

            if (err == nil) {
                // Write the response
                job.SetResponseCodeIfUndef(http.StatusOK)
                job.writeHeader()
                job.write(resp)
                job.close()
            } else {
                writeError(job, err)
            }
        }       
    }
}

// ErrorWriter is a writer used for writing error responses. It 
// expects an error object in the result field of input jobs. 
func ErrorWriter(in <-chan Job) {
    errorTemplate := "Passed in %t to ErrorWriter."
    for job := range in {
        err, ok := job.GetResult().(error)
        if (!ok) {
            job.SetResponseCode(http.StatusInternalServerError)
            writeWrongInput(job, errorTemplate)
        } else {
            writeError(job, err)
        }
        
    }
}

// StringWriter is a writer used for writing string responses. It 
// expects a string in the result field of input jobs.  
func StringWriter(in <-chan Job) {
    errorTemplate := "Passed in %t to StringWriter."
    for job := range in {
        s, ok := job.GetResult().(string)
        if (!ok) {
            writeWrongInput(job, errorTemplate)
        } else {
            job.SetResponseCodeIfUndef(http.StatusOK)
        	job.writeHeader()
            job.write([]byte(s))
            job.close()
        }   
    }
}

// JsonWriter is a writer used for writing JSON responses. It 
// expects any object in the result field of input jobs that 
// can be converted to a JSON string using the json module.
func JsonWriter(in <-chan Job) {
    for job := range in {
        js, err := json.Marshal(job.GetResult())
        if err != nil {
            writeError(job, err)
        } else {
            job.SetHeader("Content-Type", "application/json")
            job.SetResponseCodeIfUndef(http.StatusOK)
            job.writeHeader()
            job.write(js)
            job.close()
        }
    }
}

// GzipWriter is a writer used for writing responses compressed 
// using Gzip compression. It expects an io.ReadCloser object in 
// the result field of input jobs.
func GzipWriter(in <-chan Job) {
    errorTemplate := "Passed in %t to GzipWriter."
	for job := range in {
		reader, ok := job.GetResult().(io.ReadCloser)
		if (!ok) {
			writeWrongInput(job, errorTemplate)
			continue
		}

        job.SetHeader("Content-Encoding", "gzip")
        job.SetHeader("Content-Type", "application/x-gzip")
        job.SetResponseCodeIfUndef(http.StatusOK)
        job.writeHeader()
		
		gzipWriter := gzip.NewWriter(job.getResponseWriter())
		io.Copy(gzipWriter, reader)
		gzipWriter.Close()
		reader.Close()
		job.close()
	}
}

// GenericWriter is a writer used for writing generic responses.
// It expects an io.ReadCloser object in the result field of 
// input jobs.
func GenericWriter(in <-chan Job) {
    errorTemplate := "Passed in %t to GenericWriter."
	for job := range in {
		reader, ok := job.GetResult().(io.ReadCloser)
		if (!ok) {
			writeWrongInput(job, errorTemplate)
			continue
		}
		job.SetResponseCodeIfUndef(http.StatusOK)
        job.writeHeader()
		io.Copy(job.getResponseWriter(), reader)
		reader.Close()
		job.close()
	}
}

// ResponseWriter is a writer used for writing generic responses
// that were obtained from another server. It expects a
// mpserver.Response object in the result field of input jobs.
func ResponseWriter(in <-chan Job) {
    errorTemplate := "Passed in %t to ResponseWriter."
    for job := range in {
        resp, ok := job.GetResult().(Response)
        if (!ok) {
            writeWrongInput(job, errorTemplate)
            continue
        }
        // Set Headers
        header := job.getResponseWriter().Header()
        for key, value := range resp.Header {
            header.Set(key, strings.Join(value, ""))
        }

        job.SetResponseCodeIfUndef(resp.ResponseCode)
        job.writeHeader()
        job.write(resp.Body)
        job.close()
    }
}

// HttpResponseWriter is a writer used for writing generic 
// responses that were obtained from another server. It expects a
// http.Response object in the result field of input jobs.
func HttpResponseWriter(in <-chan Job) {
    errorTemplate := "Passed in %t to HttpResponseWriter."
    for job := range in {
        resp, ok := job.GetResult().(*http.Response)
        if (!ok) {
            writeWrongInput(job, errorTemplate)
            continue
        }
        // Set Headers
        header := job.getResponseWriter().Header()
        for key, value := range resp.Header {
            header.Set(key, strings.Join(value, ""))
        }
        job.writeHeader()

        func() {
            defer resp.Body.Close()
            io.Copy(job.getResponseWriter(), resp.Body)
            job.close()
        } ()
    }
}