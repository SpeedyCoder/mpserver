package main

import (
	"flag"
	"log"
    "net/http"
    "mpserver"
    "github.com/jnwhiteh/webpipes"
)

func main() {
	dir := "fileServerTest"
	useGo := flag.Bool("go", false, "use plain go implementation")
	useMPS := flag.Bool("mpserver", false, "use mpserver implementation")
	useWB := flag.Bool("webpipes", false, "use webpipes implementation")
	flag.Parse()

	if (!*useGo && !*useMPS && !*useWB){
		*useGo = true; *useMPS = true; *useWB = true
	}

	mux := http.NewServeMux()
	if (*useGo) {
		log.Println("Using plain go implementation.")
		mux.Handle("/go/", http.StripPrefix(
			"/go", http.FileServer(http.Dir(dir))))
	}

	if (*useMPS) {
		log.Println("Using mpserver implementation.")
		mux.Handle("/mpserver/", mpserver.SimpleFileServer(dir, "/mpserver"))
	}

	if (*useWB) {
		log.Println("Using webpipes implementation.")
		mux.Handle("/webpipes/", webpipes.Chain(
			webpipes.FileServer(dir, "/webpipes"),
			webpipes.OutputPipe))
	}

	log.Println("Listening on port 8080...")
    http.ListenAndServe(":8080", mux)
}



