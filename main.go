package main

import (
	"io"
	"log"
	"net/http"
)

var version string = "v0.0.1"
var myuser string = "mistermanager"
var myhome string = "/var/lib/mistermanager"
var myrepos string = myhome + "/repos"

func main() {
	http.HandleFunc("/", rootHandler)
	//http.HandleFunc("/build", buildHandler)

	log.Println("Mister Manager " + version + " Started")
	http.ListenAndServe(":8080", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Mister Manager "+version)
}
