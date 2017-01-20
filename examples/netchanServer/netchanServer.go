package main

import(
    "log"
    "net/http"
    "mpserver"
    "golang.org/x/exp/old/netchan"
)

func main() {
	expChans := mpserver.ExportChans("tcp", ":5000", []string{"in", "out"}, 
							      []netchan.Dir{netchan.Recv, netchan.Send})
	sComp := mpserver.StringComponent("Hello from the other side.")
	go sComp(expChans[0], expChans[1])

    mux := http.NewServeMux()
    impChans := mpserver.ImportChans("tcp", ":5000", []string{"in", "out"}, 
							      []netchan.Dir{netchan.Send, netchan.Recv})
    toOther := impChans[0]
    fromOther := impChans[1]

    errChan := make(mpserver.ValueChan)
    go mpserver.StringWriter(fromOther, errChan)
    go mpserver.ErrorWriter(errChan)
    mpserver.Listen(mux, "/", toOther)
    log.Println("Listening on port 3000...")
    http.ListenAndServe(":3000", mux)
}