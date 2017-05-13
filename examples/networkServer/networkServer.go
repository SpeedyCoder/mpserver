package main

import(
    "log"
    "net/http"
    "mpserver"
)

func stringer(in <-chan mpserver.Value, out chan<- mpserver.Value) {
	for val := range in {
		res, _ := val.GetResult().(mpserver.Response)
		val.SetResult(string(res.Body))
		out <- val
	}
	close(out)
}

func main() {
	//--------------------- Internal server ---------------------------
    mux := http.NewServeMux()
    in := mpserver.GetChan()
    out := mpserver.GetChan()
    sComp := mpserver.ConstantComponent("Hello world!")

    go sComp(in, out)
    go mpserver.StringWriter(out)

    mpserver.Listen(mux, "/hello", in)
    log.Println("Listening on port 3000 for internal requests...")
    go http.ListenAndServe(":3000", mux)

    //--------------------- External server ---------------------------
    in = mpserver.GetChan()
    out = mpserver.GetChan()
    toErrorWriter := mpserver.GetChan()
    toStringWriter := mpserver.GetChan()

    req, _ := http.NewRequest("GET", "http://localhost:3000/hello", nil)
    combComp := mpserver.LinkComponents(
    	mpserver.ConstantComponent(req),
        mpserver.NetworkComponent(&http.Client{}),
        mpserver.ErrorPasser(mpserver.ResponseReader),
    	mpserver.ErrorPasser(stringer))

    go mpserver.StaticLoadBalancer(combComp, 10)(in, out)
    go mpserver.ErrorRouter(out, toStringWriter, toErrorWriter)
    go mpserver.StringWriter(toStringWriter)
    go mpserver.ErrorWriter(toErrorWriter)

    mux = http.NewServeMux()
    mpserver.Listen(mux, "/", in)
    log.Println("Listening on port 5000...")
    http.ListenAndServe(":5000", mux)
}