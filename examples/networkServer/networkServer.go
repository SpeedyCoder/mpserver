package main

import(
    "log"
    "net/http"
    "mpserver"
)

func stringer(in <-chan mpserver.Value, out chan<- mpserver.Value) {
	for val := range in {
		res, _ := val.Result.(mpserver.Response)
		val.Result = string(res.Body)
		out <- val
	}
	close(out)
}

func main() {
	//--------------------- Internal server ---------------------------
    mux := http.NewServeMux()
    in := make(mpserver.ValueChan)
    out := make(mpserver.ValueChan)
    errChan := make(mpserver.ValueChan)
    sComp := mpserver.ConstantComponent("Hello world!")

    go sComp(in, out)
    go mpserver.StringWriter(out, errChan)
    go mpserver.ErrorWriter(errChan)

    mpserver.Listen(mux, "/hello", in)
    log.Println("Listening on port 3000 for internal requests...")
    go http.ListenAndServe(":3000", mux)

    //--------------------- External server ---------------------------
    in = make(mpserver.ValueChan)
    out = make(mpserver.ValueChan)
    errChan = make(mpserver.ValueChan)

    req, _ := http.NewRequest("GET", "http://localhost:3000/hello", nil)
    combComp := mpserver.LinkComponents(
    	mpserver.ConstantComponent(req),
    	mpserver.NetworkComponent(&http.Client{}),
    	mpserver.ResponseProcessor,
    	mpserver.ErrorPasser(stringer))

    go combComp(in, out)
    go mpserver.AddErrorSplitter(mpserver.StringWriter)(out, errChan)
    go mpserver.ErrorWriter(errChan)

    mux = http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 5000...")
    http.ListenAndServe(":5000", mux)
}