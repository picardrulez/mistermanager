package main

import (
	"flag"
	"io"
	"log"
	"net/http"
)

var version string = "v0.1.1"
var myuser string = "root"
var myhome string = "/var/lib/mistermanager"
var myrepos string = myhome + "/repos"

func main() {
	var config = ReadConfig()

	//Handling user flags
	bind := flag.String("bind", config.Bind, "port to bind to")
	flag.Parse()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/build", buildHandler)

	log.Println("Mister Manager " + version + " Started")
	http.ListenAndServe(":"+*bind, nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Mister Manager "+version)
}
