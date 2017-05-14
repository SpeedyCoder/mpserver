package main
import "mpserver"

func main() {
	// Create the channels
    in := mpserver.GetChan()
    out := mpserver.GetChan()

    // Start the components
    mpserver.Listen("/", in, nil)
    go mpserver.ConstantComponent("Hello world!")(in, out)
    go mpserver.StringWriter(out)
    
    mpserver.ListenAndServe(":3000", nil) // Start the server
}