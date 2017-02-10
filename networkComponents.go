package mpserver

import (
	"log"
	"net/http"
	"io/ioutil"
	"errors"
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
			go func () {
				resp, err := client.Do(req)
				if err != nil {
					val.Result = err
				} else {
					val.Result = resp
				}
				out <- val
			}()	
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

// Can be removed
// func BodyExtractor(in <-chan Value, out chan<- Value) {
// 	for val := range in {
// 		resp, ok := val.Result.(*http.Response)
// 		if !ok {
// 			val.Result = errors.New("No response provided to Body Extractor Component.")
// 			out <- val
// 			continue
// 		}

// 		body, err := readResponse(resp)
// 		if err != nil {
// 			val.Result = err
// 		} else {
// 			val.Result = body
// 		}
// 		out <- val

// 	}
// 	close(out)
// }

func RequestCopier(scheme, host string) Component {
	return func (in <-chan Value, out chan<- Value) {
		for val := range in {
			val.Request.URL.Scheme = scheme
			val.Request.URL.Host = host
			log.Println(val.Request.URL)
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
			val.ResponseCode = resp.StatusCode
			val.Result = Response{resp.Header, body}
		}
		out <- val

	}
	close(out)
}

func ProxyComponent(scheme, host string, client *http.Client) Component {
	return LinkComponents(
		ErrorPasser(RequestCopier(scheme, host)),
		ErrorPasser(NetworkComponent(client)),
		ErrorPasser(ResponseProcessor))
}



