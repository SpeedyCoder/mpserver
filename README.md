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
* For every value that a Component reads from the input channel, 
it should write a value to the output channel. So, that every request 
gets a response.

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
When a condition returns true the value is written to a corresponding output channel 
and the processing of the current value terminates. If all conditions return false
then the value is written to the default output channel.

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
id to the next state, which is generated from the initial state using the provided value,
is stored. The result of the call to the `Result` function on the next state with the provided 
value is then sent down the pipeline.

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

Cache Component is a Component generator with the following signature:
```go
func CacheComponent(worker Component, expiration time.Duration) Component
```
The returned component behaves as follows:
* When a value is inputted that the component hasn't seen before, then
the value is passed to the provided worker and the result is stored in
a map.
* When the component gets a previously seen value, it returns the value
stored in the map, if it hasn't expired yet. If it expired, then the component
treats it as a new value.
* Expired values are regularly deleted from the map.

For the generated component to function properly, the worker must output
a value for every value that is sent to it.

## Load Balancing

Load balancing component a Component generator with the following signature:
```go
func LoadBalancingComponent(addTimeout, removeTimeout time.Duration, worker Component) Component
```
The returned component has an array of workers which can be easily shut down.
The initial number of workers is 1. 

For every inputted value, the component behaves as follows:
* It tries to send the value to the workers, if this is not successful, 
before the addTimeout, then the component creates a new worker and then
tries to send the value to the workers again. This will be successful as
there is at least one worker that is not busy. 
* The workers send their results further down the pipeline.
* If there are no incoming values for removeTimout time, then the component
shuts down one of the workers, if there is more than one worker.
* When the input channel of the component is closed, then it shuts down all
the workers, then it closes its output channel and terminates.

## Error handling

Error Passer is a component generator with the following signature:
```go
func ErrorPasser(worker Component) Component
```
The generated component behaves as follows:
* For every inputted value it checks whether the result is an error
and if that is the cases it just outputs the value.
* If the result of the inputted value is not an error, then the component
passes the value to the worker, which after processing it passes 
it further down the pipeline.

Panic handling component is a component generator with the following signature:
```go
func PannicHandlingComponent(worker Component) Component
```
It takes a component that can cause panic and returns a component the behaves as follows:
* It passes every inputted value to the worker and gets the result from it and then passes
the result further down the pipeline.
* In case the worker crashes, that is causes panic, the component writes error as a result 
for the value that caused the crash and the restarts the worker.
* The component can itself cause a panic if the worker closes its input channel, or if 
it closes its output channel before its input channel is closed. That is if the worker
violates some of the basic requirements for Components.

## Network Components

TODO


