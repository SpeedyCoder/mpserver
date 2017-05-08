package main
import (
	"log"
	"net/http"
	"mpserver"
	"time"
)

func main() {
	// Define the components
	proxy := mpserver.ProxyComponent("http", "www.google.co.uk",
			    &http.Client{}, time.Second, time.Second*60, 100)
	store := mpserver.NewMemStore()
	cachedComp := mpserver.CacheComponent(
					store, proxy, time.Second*20, true)

	// Create channels
	in := mpserver.GetChan()
	out := mpserver.GetChan()
	toWriter := mpserver.GetChan()
	errChan := mpserver.GetChan()

	// Start the components and writers
	go cachedComp(in, out)
	go mpserver.ErrorSplitter(out, toWriter, errChan)
	go mpserver.ResponseWriter(toWriter)
	go mpserver.ErrorWriter(errChan)

	// Start the server
	mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 5000...")
    http.ListenAndServe(":5000", mux)
}