package main

import (
	"log"
	"net/http"
	"mpserver"
)

func main() {
	comp := mpserver.LinkComponents(
		mpserver.ErrorPasser(mpserver.RequestCopier("http", "www.google.co.uk")),
		mpserver.ErrorPasser(mpserver.NetworkComponent(&http.Client{})),
		mpserver.ErrorPasser(mpserver.ResponseProcessor))

	in := make(mpserver.ValueChan)
	toSplitter := make(mpserver.ValueChan)
	out := make(mpserver.ValueChan)
	errChan := make(mpserver.ValueChan)
	go comp(in, toSplitter)
	go mpserver.ErrorSplitter(toSplitter, out, errChan)
	go mpserver.ResponseWriter(out, errChan)
	go mpserver.ErrorWriter(errChan)

	mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 5000...")
    http.ListenAndServe(":5000", mux)
}