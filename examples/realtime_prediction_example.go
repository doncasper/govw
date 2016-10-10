package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DonCasper/govw"
)

var vw govw.VWDaemon

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
	vw = govw.NewDaemon("vw", [2]int{26542, 26543}, 30, "/full/path/to/some.model", true, true)
	if err := vw.Run(); err != nil {
		log.Fatal("Error while running VW daemon!", err)
	}

	if vw.Model.Updatable {
		go govw.ModelFileChecker(&vw)
	}

	runServer()
}
