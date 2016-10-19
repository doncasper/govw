package main

import (
	"encoding/json"
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

		p, err := vw.Predict(string(body))
		if err != nil {
			log.Fatal("Error while getting predict:", err)
		}

		res, err := json.Marshal(p)
		if err != nil {
			log.Fatal("Error while marshaling prediction result:", err)
		}

		fmt.Fprintf(w, "%s", res)
	default:
		w.WriteHeader(404)
		fmt.Fprint(w, "Method unavailable!")
	}
}

func runServer() {
	addr := ":8080"
	log.Println("Server running on port", addr)

	http.HandleFunc("/", predictHandler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	//vw = govw.NewDaemon("vw", [2]int{26542, 26543}, 1000, "/full/path/to/some.model", true, true)
	vw = govw.NewDaemon("vw", [2]int{26542, 26543}, 1000, "/home/casper/Golab/src/smartyads/go_app_server/models/click.model", true, true)
	if err := vw.Run(); err != nil {
		log.Fatal("Error while running VW daemon!", err)
	}

	if vw.Model.Updatable {
		go govw.ModelFileChecker(&vw)
	}

	runServer()
}
