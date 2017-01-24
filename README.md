# mpserver

mpserver is a go library that implements a framework for building 
concurrent web-servers based on message passing.

The web-server consists of a multiple Components which are plugged 
together using channels to form a pipeline that transforms requests 
to responses.

### Values

Channels between Components transfer values of the following type:

```go
type Value struct {
    Request *http.Request
    Writer http.ResponseWriter
    Result Any
    Done chan<- bool
    ResponseCode int
}
```
Where:
* `Request` is the http request made by the client.
* `Writer` is an object that writes the response back to the client.
* `Result` is the result computed so far (initially `nil`).
* `Done` is a signaling channel. Signal should be send on this channel,
after the response is written at the end of the pipeline.
* `ResponseCode` is the response code that should be returned to the client.

### Components

Component is defined as follows:

```go
type Component func (in <-chan Value, out chan<- Value)
```
Every Component should satisfy the following:
* When run Component should read values from the input channel
and output processed values on the output channel while the input 
channel is open.
* Component terminates normally only if its input channel is closed.
* Before terminating, every Component closes its output channel.
* Component can only close its output channel and it can only do so
after its input channel have been closed.