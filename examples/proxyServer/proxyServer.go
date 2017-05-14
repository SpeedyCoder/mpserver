package main
import (
	"mpserver"
	"time"
	"net/http"
)

const CacheTimeout = time.Minute
const AddTimeout = time.Second*5
const RemoveTimeout = time.Minute

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
		// Start 10 instances of the Proxy Component
		loadProxy := mpserver.StaticLoadBalancer(proxy, 10)
		// Cache the output
		cachedProxy := mpserver.CacheComponent(
						storage, loadProxy, CacheTimeout)

		out := mpserver.GetChan()
		go cachedProxy(in, out)
		mpserver.StaticLoadBalancerWriter(writers, 10)(out)
	}
}

func main() {
	storage := mpserver.NewMemStorage()
	server := mpserver.DynamicLoadBalancerWriter(
				proxyServerWriter(storage), 10, 
				AddTimeout, RemoveTimeout)

	in := mpserver.GetChan()
	go server(in)
	go mpserver.StorageCleaner(storage, nil, CacheTimeout)

	// Start the server
    mpserver.Listen("/", in, nil)
    mpserver.ListenAndServe(":5000", nil)
}