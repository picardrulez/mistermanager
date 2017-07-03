package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

//Set global vars
var version string = "v0.1.8.2"
var logfile string = "/var/log/mistermanager"
var myuser string = "root"
var myhome string = "/var/lib/mistermanager"
var myrepos string = myhome + "/repos"

func main() {
	//Set Up Logging
	var _, err = os.Stat(logfile)
	if os.IsNotExist(err) {
		var file, err = os.Create(logfile)
		checkError(err)
		defer file.Close()
	}
	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND, 0644)
	checkError(err)
	defer f.Close()
	log.SetOutput(f)

	//Read config
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

func checkError(err error) {
	if err != nil {
		log.Println(err.Error)
	}
}
