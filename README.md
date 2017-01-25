# mpserver

mpserver is a go library that implements a framework for building 
concurrent web-servers based on message passing.

The web-server consists of a multiple Components which are plugged 
together using channels to form a pipeline that transforms requests 
to responses.

## Values

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

## Components

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

## Writers

Writer is defined as follows:

```go
type Writer func (in <-chan Value, errChan chan<- Value)
```
It is a function with input channel and channel for reporting errors.
* When a Writer writes the result to the client, it should send a signal
on done channel, so that the connection to client can be closed.
* The writer terminates when its input channel is closed. It shouldn't
close the error channel, as other processes might be using it.

#### Error Writer

Error Writer is a function with the following signature:
```go
func ErrorWriter(in <-chan Value)
```
It has only one input channel. The writer terminates when the channel
is closed. The channel should be closed only after all Writers that 
are using it terminated.


