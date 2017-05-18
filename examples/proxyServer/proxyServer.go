package main
import (
	"mpserver"
	"time"
	"net/http"
)

const CacheTimeout = time.Minute
const AddTimeout = time.Second*5
const RemoveTimeout = time.Minute
const n = 4
const k = 4

func writers(in <-chan mpserver.Job) {
	toRespWriter := mpserver.GetChan()
	toErrWriter := mpserver.GetChan()
	go mpserver.ErrorRouter(in, toRespWriter, toErrWriter)
	go mpserver.ResponseWriter(toRespWriter)
	mpserver.ErrorWriter(toErrWriter)
}

func proxyServerWriter(storage mpserver.Storage) mpserver.Writer{      
	return func (in <-chan mpserver.Job) {
		// Define the components
		proxy := mpserver.ProxyComponent("http", 
					"www.google.co.uk", &http.Client{})
		// Start n instances of the Proxy Component
		loadProxy := mpserver.StaticLoadBalancer(proxy, n)
		// Cache the output
		cachedProxy := mpserver.CacheComponent(
						storage, loadProxy, CacheTimeout)

		out := mpserver.GetChan()
		go cachedProxy(in, out)
		mpserver.StaticLoadBalancerWriter(writers, n)(out)
	}
}

func main() {
	storage := mpserver.NewMemStorage()
	server := mpserver.DynamicLoadBalancerWriter(
				proxyServerWriter(storage), k, 
				AddTimeout, RemoveTimeout)

	in := mpserver.GetChan()
	go server(in)
	go mpserver.StorageCleaner(storage, nil, CacheTimeout)

	// Start the server
    mpserver.Listen("/", in, nil)
    mpserver.ListenAndServe(":5000", nil)
}