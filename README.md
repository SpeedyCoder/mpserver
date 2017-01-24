# mpserver

mpserver is a go library that implements a framework for building 
concurrent web-servers based on message passing.

The web-server consists of a multiple Components which are plugged 
together using channels to form a pipeline that transforms requests 
to responses.

## Components

Component is defined as follows:

```go
type Component func (in <-chan Value, out chan<- Value)
```
That is, it is a function that takes an input and output channel.

Every should behave as follows:
* When run Component should read values from the input channel
and output processed values on the output channel while the input 
channel is open.
* Component terminates normally only if its input channel is closed.
* Before terminating, every Component closes its output channel.

