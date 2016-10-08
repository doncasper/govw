package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DonCasper/govw"
)

const msg = "UNIQ-ID| 100:1 200:1 300:1 400:1 500: 600:1 700:1"

var vw *govw.VWDaemon

func predictHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal("Error while reading request body:", err)
		}

		p, err := vw.Predict(body)
		if err != nil {
			log.Fatal("Error while getting predict!")
		}
		fmt.Fprint(w, p)
	default:
		w.WriteHeader(404)
		fmt.Fprint(w, "Method unavailable!")
	}
}

func runServer() {
	addr := ":8080"
	log.Println("Server running on port", addr)

	http.HandleFunc("/", predictHandler)
	http.ListenAndServe(addr, nil)
}

func main() {
	//vw = govw.NewDaemon("vw", 26542, 30, "/full/path/to/some.model", true, false)
	vw = govw.NewDaemon("vw", 26542, 30, "/home/casper/Golab/src/smartyads/go_app_server/models/click.model", true, false)
	vw.Run()

	runServer()
}
