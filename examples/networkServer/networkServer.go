package main

import(
    "log"
    "net/http"
    "mpserver"
)

func stringer(in <-chan mpserver.Value, out chan<- mpserver.Value) {
	for val := range in {
		res, _ := val.Result.([]byte)
		val.Result = string(res)
		out <- val
	}
	close(out)
}

func main() {
	// Internal server
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)
    sComp := mpserver.StringComponent("Hello world!")
    go sComp(in, out)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/hello", in)
    log.Println("Listening on port 3000 for internal requests...")
    go http.ListenAndServe(":3000", mux)

    // External server
    in = make(mpserver.ValueChan)
    toNetComp := make(mpserver.ValueChan)
    toStringer := make(mpserver.ValueChan)
    toSplitter := make(mpserver.ValueChan)
    out = make(mpserver.ValueChan)
    errChan = make(mpserver.ValueChan)

    req, _ := http.NewRequest("GET", "http://localhost:3000/hello", nil)
    go mpserver.ConstantComponent(req)(in, toNetComp)
    netComp := mpserver.NetworkComponent(&http.Client{})
    go netComp(toNetComp, toStringer)
    str := mpserver.ErrorPassingComponent(stringer)
    go str(toStringer, toSplitter)
    go mpserver.ErrorSplitter(toSplitter, out, errChan)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)
    mux = http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 5000...")
    http.ListenAndServe(":5000", mux)

}