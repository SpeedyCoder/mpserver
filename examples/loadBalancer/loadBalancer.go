package main

import (
	"net/http"
	"mpserver"
	"time"
	"log"
)

func main() {
	worker := mpserver.MakeComponent(
		func (val mpserver.Value) {
			time.Sleep(time.Second*5)
			val.SetResult("Hello World!")
	})

	lb := mpserver.DynamicLoadBalancer(
			worker, 10, time.Second, time.Second*60, )
	// lb := mpserver.StaticLoadBalancer(worker, 10)

	in := mpserver.GetChan()
	out := mpserver.GetChan()
	go lb(in, out)
	go mpserver.StringWriter(out)

	mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}