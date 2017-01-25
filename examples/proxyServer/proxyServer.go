package main

import (
	"log"
	"net/http"
	"mpserver"
	"time"
)

func main() {
	proxy := mpserver.ProxyComponent("http", "www.google.co.uk",&http.Client{})
	cachedComp := mpserver.CacheComponent(proxy, time.Second*60)

	in := make(mpserver.ValueChan)
	out := make(mpserver.ValueChan)
	errChan := make(mpserver.ValueChan)

	go cachedComp(in, out)
	go mpserver.AddErrorSplitter(mpserver.ResponseWriter)(out, errChan)
	go mpserver.ErrorWriter(errChan)

	mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 5000...")
    http.ListenAndServe(":5000", mux)
}