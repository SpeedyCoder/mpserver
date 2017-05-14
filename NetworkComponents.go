package mpserver

import (
	"net/http"
	"io/ioutil"
	"errors"
)

// NetworkComponent generates a component that uses the provided
// client to perform http.Request, that is taken from the 
// result field of the input jobs. The generated component
// writes the obtained http.Response or an error to the result 
// field of the job before outputting it. Components further
// down the network are responsible for closing the body of the
// response.
// 
// Client is safe for concurrent use, so the returned component 
// can be used multiple times.
func NetworkComponent(client *http.Client) Component {
	inputError := errors.New(
		"No request provided to Network Component.")
	return func (in <-chan Job, out chan<- Job) {
		for job := range in {
			req, ok := job.GetResult().(*http.Request)
			if !ok {
				job.SetResult(inputError)
				out <- job
				continue
			}
			// Perform the request using the client.
			resp, err := client.Do(req)
			if err != nil {
				// Request wasn't successful.
				job.SetResult(err)
			} else {
				// Request was successful.
				job.SetResult(resp)
			}
			out <- job
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
// in the request field of input jobs with the provided host 
// and scheme and then stores it in the result field of the job
// before outputting it.
func RequestRewriter(scheme, host string) Component {
	return func (in <-chan Job, out chan<- Job) {
		for job := range in {
			request := job.GetRequest()
			// Overwrite the scheme and host.
			request.URL.Scheme = scheme
			request.URL.Host = host

			// Clean the RequestURI and Host fields to be able to
			// make perform the Request again.
			request.RequestURI = ""
			request.Host = ""
			job.SetResult(request)
			out <- job
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
// an input job. Then it creates a Response object using the 
// data that was read and writes it to the result field of the 
// job before outputting it.
func ResponseReader(in <-chan Job, out chan<- Job) {
	inputError := errors.New(
		"No http response provided to ResponseReader.")
	for job := range in {
		resp, ok := job.GetResult().(*http.Response)
		if !ok {
			job.SetResult(inputError)
			out <- job
			continue
		}

		body, err := readResponse(resp)
		if err != nil {
			job.SetResult(err)
		} else {
			job.SetResponseCode(resp.StatusCode)
			job.SetResult(
				Response{resp.Header, resp.StatusCode, body})
		}
		out <- job

	}
	close(out)
}

// ProxyComponent returns a component that forwards all requests
// in the request field of all input jobs to the provided host 
// using the specified scheme and client. The output jobs  
// contain the corresponding Response object in the result field.
func ProxyComponent(scheme, host string, 
					client *http.Client) Component {
	return LinkComponents(
		ErrorPasser(RequestRewriter(scheme, host)),
		ErrorPasser(NetworkComponent(client)),
		ErrorPasser(ResponseReader))
}