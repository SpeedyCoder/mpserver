package mpserver

import (
    "net/http"
    "golang.org/x/net/websocket"
)

// Job is an interface that represents the type of messages 
// that are passed between the components of the network. It 
// stores the client request and the result computed so far.
type Job interface {
    // GetRequest returns a pointer to the http request made by 
    // the client 
    GetRequest() *http.Request

    // GetResult returns the current value in the result field.
    // The value is initially set to nil
    GetResult() interface{}

    // SetResult sets the result field to the provided value
    SetResult(value interface{})

    // SetResponseCode sets the response code of this value
    SetResponseCode(respCode int)

    // SetResponseCode sets the response code of this value if
    // it is undefined
    SetResponseCodeIfUndef(respCode int)

    // SetHeader sets the the header of the response with the
    // provided key and value
    SetHeader(key string, value string)

    // Private methods
    getResponseWriter() http.ResponseWriter
    getResponseCode() int
    writeHeader()
    write([]byte)
    close()
}

type jobStruct struct {
    request *http.Request
    result interface{}

    responseCode int
    responseWriter http.ResponseWriter
    webSocket *websocket.Conn
    done chan<- bool
}

func (job jobStruct) GetRequest() *http.Request {
    return job.request
}

func (job jobStruct) GetResult() interface{} {
    return job.result
}

func (job *jobStruct) SetResult(newResult interface{}) {
    job.result = newResult
}

func (job *jobStruct) SetResponseCode(responseCode int) {
    job.responseCode = responseCode;
}

func (job *jobStruct) SetResponseCodeIfUndef(responseCode int) {
    if (job.responseCode == UndefinedRespCode) {
        job.responseCode = responseCode;
    }
}

func (job jobStruct) SetHeader(key, value string) {
    job.responseWriter.Header().Set(key, value)
}

func (job jobStruct) getResponseWriter() http.ResponseWriter {
    return job.responseWriter
}

func (job jobStruct) getResponseCode() int {
    return job.responseCode
}

func (job jobStruct) writeHeader() {
    job.responseWriter.WriteHeader(job.responseCode);
}

func (job jobStruct) write(body []byte) {
    job.responseWriter.Write(body)
}

func (job jobStruct) close() {
    job.done <- true;
}