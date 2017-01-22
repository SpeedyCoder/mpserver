package mpserver

import (
	"log"
	"net/http"
	"io/ioutil"
	"errors"
)

func doRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return []byte{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func NetworkComponent(client *http.Client) Component {
	// Client is safe for concurrent use, so the returned component can
	// be used multiple times
	return func (in <-chan Value, out chan<- Value) {
		for val := range in {
			req, ok := val.Result.(*http.Request)
			if !ok {
				val.Result = errors.New("No request provided to Network Component.")
				out <- val
				continue
			}

			body, err := client.Do(req)
			if err != nil {
				val.Result = err
			} else {
				val.Result = body
				log.Println(body)
			}
			out <- val
		}
		close(out)
	}
}

