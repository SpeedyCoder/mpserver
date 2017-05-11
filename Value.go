package mpserver

import (
    "net/http"
    "golang.org/x/net/websocket"
)

// Value is an interface that represents the type of messages 
// that are passed between the components of the network. 
type Value interface {
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

type valueStruct struct {
    request *http.Request
    result interface{}

    responseCode int
    responseWriter http.ResponseWriter
    webSocket *websocket.Conn
    done chan<- bool
}

func (val valueStruct) GetRequest() *http.Request {
    return val.request
}

func (val valueStruct) GetResult() interface{} {
    return val.result
}

func (val *valueStruct) SetResult(newResult interface{}) {
    val.result = newResult
}

func (val *valueStruct) SetResponseCode(responseCode int) {
    val.responseCode = responseCode;
}

func (val *valueStruct) SetResponseCodeIfUndef(responseCode int) {
    if (val.responseCode == UndefinedRespCode) {
        val.responseCode = responseCode;
    }
}

func (val valueStruct) SetHeader(key, value string) {
    val.responseWriter.Header().Set(key, value)
}

func (val valueStruct) getResponseWriter() http.ResponseWriter {
    return val.responseWriter
}

func (val valueStruct) getResponseCode() int {
    return val.responseCode
}

func (val valueStruct) writeHeader() {
    val.responseWriter.WriteHeader(val.responseCode);
}

func (val valueStruct) write(body []byte) {
    val.responseWriter.Write(body)
}

func (val valueStruct) close() {
    val.done <- true;
}