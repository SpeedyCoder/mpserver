package main
import "net/http"
import "mpserver"

func main() {
    in := mpserver.GetChan()
    out := mpserver.GetChan()

    mux := http.NewServeMux()
    mpserver.Listen(mux, "/", in)

    go mpserver.ConstantComponent("Hello world!")(in, out)
    go mpserver.StringWriter(out)
    
    http.ListenAndServe(":3000", mux)
}