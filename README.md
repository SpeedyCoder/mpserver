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

Component is a generic part of the pipeline with an input and output channel.
The type Component is defined as follows:

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

Writer is the end of the pipeline which writes results to the client.
The type Writer is defined as follows:

```go
type Writer func (in <-chan Value, errChan chan<- Value)
```
It is a function with input channel and channel for reporting errors.
* When a Writer writes the result to the client, it should send a signal
on done channel, so that the connection to client can be closed.
* The writer terminates when its input channel is closed. It shouldn't
close the error channel, as other processes might be using it.

### Error Writer

Error Writer is a function with the following signature:
```go
func ErrorWriter(in <-chan Value)
```
It has only one input channel. The writer terminates when the channel
is closed. The channel should be closed only after all Writers that 
are using it terminated.

## Conditions and Splitter

Condition type is defined as follows:

```go
type Condition func (val Value) bool
```
It is a function that takes a value and returns a boolean.

Conditions are used in Splitter, which is a component with the 
following signature:
```go
func Splitter(in <-chan Value, defOut chan<- Value, outs []chan<- Value, conds []Condition)
```
It takes an input channel, default output channel, array of output channels and array of conditions.
The number of output channels and the number of conditions should be the same.
For every input value the Splitter evaluates the conditions from first to last.
When a condition returns true the value is written to a corresponding output channel.
If all conditions return false then the value is written to the default output channel.

## States and Session Management Component

State is an interface defined as follows:
```go
type State interface {
    Next(val Value) (State, error)
    Terminal() bool
    Result() Any
}
```
* The `Next` function returns the next state when provided with a value.
* The `Terminal` function indicates whether the current state is a terminal state.
* The `Result function` function returns the result that possibly after further 
processing should be returned to the client.

Session management component is a following component generator.
```go
func SessionManagementComponent(initial State, seshExp time.Duration) Component
```
It takes an initial state for each session and a session expiration time and 
returns a Component that behaves as a session manager. 

The component behaves as follows:

The component stores a mapping from session ids to current state of the session.

Every new request gets assigned a unique session id and a mapping from the generated 
id to the initial state is stored. The result of the call to the `Result` function
on the initial state with the provided value is then sent down the pipeline.

When a request with set `Session-Id` header comes in, the component gets the current 
state for the session and gets the next state for it based on the provided value.
The mapping is then updated with the generated state and the result from the current state
is passed down the pipeline.

When a session reaches a Terminal state, the session id is removed from the map
and the final result is outputted.

The states in the mapping are timestamped and are removed from the map after
the session expiration time, if they haven't finished before that.

The abstraction using states allows the users of the package to implement the
logic for the session, independently of the session management part.

## Caching

TODO

## Load Balancing

TODO

## Error handling

TODO

## Network Components

TODO


