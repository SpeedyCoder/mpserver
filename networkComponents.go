package mpserver

import (
	"log"
	"net/http"
	"io/ioutil"
	"errors"
	"time"
)

func NetworkComponent(client *http.Client) Component {
	/** Client is safe for concurrent use, so the returned component can
	  * be used multiple times.
	  * Outputs a response and components further down the chain are 
	  * responsible for closing the response body*/
	return func (in <-chan Value, out chan<- Value) {
		for val := range in {
			req, ok := val.Result.(*http.Request)
			if !ok {
				val.Result = errors.New("No request provided to Network Component.")
				out <- val
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				val.Result = err
			} else {
				val.Result = resp
			}
			out <- val
		}
		close(out)
	}
}

// Helper function to make and process the request
func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func RequestCopier(scheme, host string) Component {
	return func (in <-chan Value, out chan<- Value) {
		for val := range in {
			val.Request.URL.Scheme = scheme
			val.Request.URL.Host = host
			log.Println("Request Copier", val.Request.URL)
			val.Request.RequestURI = ""
			val.Request.Host = ""
			val.Result = val.Request
			out <- val
		}
		close(out)
	}
}

type Response struct {
	Header http.Header
	ResponseCode int
	Body []byte
}

func ResponseProcessor(in <-chan Value, out chan<- Value) {
	for val := range in {
		resp, ok := val.Result.(*http.Response)
		if !ok {
			val.Result = errors.New("No response provided to to Response Processor Component.")
			out <- val
			continue
		}

		body, err := readResponse(resp)
		if err != nil {
			val.Result = err
		} else {
			val.SetResponseCode(resp.StatusCode)
			val.Result = Response{resp.Header, resp.StatusCode, body}
		}
		out <- val

	}
	close(out)
}

func ProxyComponent(scheme, host string, client *http.Client, 
		addTimeout, removeTimeout time.Duration, nReq int) Component {
	return LinkComponents(
		ErrorPasser(RequestCopier(scheme, host)),
		DynamicLoadBalancer(
				addTimeout, removeTimeout,
				LinkComponents(
					ErrorPasser(NetworkComponent(client)),
					ErrorPasser(ResponseProcessor)),
				nReq))
}

