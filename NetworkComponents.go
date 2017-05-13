package mpserver

import (
	"net/http"
	"io/ioutil"
	"errors"
)

// NetworkComponent generates a component that uses the provided
// client to perform http.Request, that is taken from the 
// result field of the input values. The generated component
// writes the obtained http.Response or an error to the result 
// field of the value before outputting it. Components further
// down the network are responsible for closing the body of the
// response.
// 
// Client is safe for concurrent use, so the returned component 
// can be used multiple times.
func NetworkComponent(client *http.Client) Component {
	inputError := errors.New(
		"No request provided to Network Component.")
	return func (in <-chan Value, out chan<- Value) {
		for val := range in {
			req, ok := val.GetResult().(*http.Request)
			if !ok {
				val.SetResult(inputError)
				out <- val
				continue
			}
			// Perform the request using the client.
			resp, err := client.Do(req)
			if err != nil {
				// Request wasn't successful.
				val.SetResult(err)
			} else {
				// Request was successful.
				val.SetResult(resp)
			}
			out <- val
		}
		close(out)
	}
}

// readResponse is a helper function read the body of 
// http.Response.
func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}


// RequestRewriter returns a component that modifies the request 
// in the request field of input values with the provided host 
// and scheme and then stores it in the result field of the value
// before outputting it.
func RequestRewriter(scheme, host string) Component {
	return func (in <-chan Value, out chan<- Value) {
		for val := range in {
			request := val.GetRequest()
			// Overwrite the scheme and host.
			request.URL.Scheme = scheme
			request.URL.Host = host

			// Clean the RequestURI and Host fields to be able to
			// make perform the Request again.
			request.RequestURI = ""
			request.Host = ""
			val.SetResult(request)
			out <- val
		}
		close(out)
	}
}

// Response is a type that represents an http response that has
// been read.
type Response struct {
	Header http.Header // Headers of the response.
	ResponseCode int   // Response code of the response.
	Body []byte		   // Body of the response.
}

// ResponseReader is a component that reads the body of the 
// provided http.Response that is taken from the result field of 
// an input value. Then it creates a Response object using the 
// data that was read and writes it to the result field of the 
// value before outputting it.
func ResponseReader(in <-chan Value, out chan<- Value) {
	inputError := errors.New(
		"No http response provided to ResponseReader.")
	for val := range in {
		resp, ok := val.GetResult().(*http.Response)
		if !ok {
			val.SetResult(inputError)
			out <- val
			continue
		}

		body, err := readResponse(resp)
		if err != nil {
			val.SetResult(err)
		} else {
			val.SetResponseCode(resp.StatusCode)
			val.SetResult(Response{resp.Header, resp.StatusCode, body})
		}
		out <- val

	}
	close(out)
}

// ProxyComponent returns a component that forwards all requests
// in the request field of all input values to the provided host 
// using the specified scheme and client. The output values  
// contain the corresponding Response object in the result field.
func ProxyComponent(scheme, host string, 
					client *http.Client) Component {
	return LinkComponents(
		ErrorPasser(HTTPRequestMaker(scheme, host)),
		ErrorPasser(NetworkComponent(client)),
		ErrorPasser(ResponseReader))
}