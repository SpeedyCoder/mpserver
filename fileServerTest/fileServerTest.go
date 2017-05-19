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
	useGo := flag.Bool(
		"go", false, "use plain go implementation")
	useMPS := flag.Int(
		"mpserver", 0, "use one of mpserver implementations")
	useWB := flag.Bool(
		"webpipes", false, "use webpipes implementation")
	flag.Parse()

	if (!*useGo && *useMPS == 0 && !*useWB){
		*useGo = true; *useMPS = 1; *useWB = true
	}

	mux := http.NewServeMux()
	if (*useGo) {
		log.Println("Using plain go implementation.")
		mux.Handle("/go/", http.StripPrefix(
			"/go", http.FileServer(http.Dir(dir))))
	}
	if (*useMPS != 0) {
		log.Printf("Using mpserver implementation %v.", *useMPS)
		switch *useMPS {
			case 1: {
				mux.Handle("/mpserver/", 
					mpserver.SimpleFileServer(dir, "/mpserver"))
			}
			case 2: {
				mux.Handle("/mpserver/", 
					mpserver.SBalancedFileServer(
						dir, "/mpserver", 4))
			}
			case 3: {
				mux.Handle("/mpserver/", 
					mpserver.DBalancedFileServer(
						dir, "/mpserver", 4))
			}
			default: {
				log.Fatal(
					"Allowed values for this flag are 1, 2, 3")
			}
		}
		
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