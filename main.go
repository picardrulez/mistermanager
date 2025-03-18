package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

//Set global vars
var version string = "v1.1.0"
var logfile string = "/var/log/mistermanager"
var myuser string = "root"
var myhome string = "/var/lib/mistermanager"
var myrepos string = myhome + "/repos"

type readStruct struct {
	Version string `json:"version"`
}

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
	http.HandleFunc("/versions", versionHandler)

	log.Println("Mister Manager " + version + " Started")
	http.ListenAndServe(":"+*bind, nil)
}

// http handler "/" root function
func rootHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Mister Manager "+version)
}

// http handler /versions  displays all versions of managed app on slaves
func versionHandler(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("error getting hostname")
	}
	var config = ReadConfig()
	//get list of managed slaves
	notifyManagers := config.Managers
	//get version path
	versionPath := config.VersionPath
	//get version format
	versionFormat := config.VersionFormat
	stripQuotes := strings.Replace(versionPath, "\\", "", -1)
	//build local url to request version
	localURL := "http://" + hostname + stripQuotes
	//get local managed app version
	response, err := http.Get(localURL)
	if err != nil {
		log.Println("error pulling version from " + hostname)
		log.Printf("%s", err)
	}
	if versionFormat == "json" {
		decoder := json.NewDecoder(response.Body)
		var jsonResponse readStruct
		err = decoder.Decode(&jsonResponse)
		if err != nil {
			log.Println("error decoding json")
		}
		io.WriteString(w, hostname+" "+jsonResponse.Version+"\n")
	} else {
		pageContent, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("error reading version page contents from " + hostname)
			log.Printf("%s", err)
			stringPageReturn := string(pageContent)
			io.WriteString(w, hostname+" "+stringPageReturn+"\n")
		}
	}
	var versionchan chan string = make(chan string)
	//if local box is a master, ask for versions from slaves
	if len(notifyManagers) != 0 {
		for i := 0; i < len(notifyManagers); i++ {
			notifyBox := notifyManagers[i]
			stripQuotes := strings.Replace(versionPath, "\\", "", -1)
			versionURL := "http://" + notifyBox + stripQuotes
			go managedVersion(notifyBox, versionURL, versionchan, versionFormat)
		}
		var versionArray []string
		for i := 0; i < len(notifyManagers); i++ {
			versionReturn := <-versionchan
			versionArray = append(versionArray, versionReturn)
		}
		sort.Strings(versionArray)
		for i := 0; i < len(notifyManagers); i++ {
			io.WriteString(w, versionArray[i])
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

//get local managed app version
func managedVersion(notifyBox string, url string, versionchan chan string, format string) int {
	//makes http request for version
	response, err := http.Get(url)
	if err != nil {
		log.Println("error pulling version from " + notifyBox)
		log.Printf("%s", err)
		return 2
	}
	if format == "json" {
		decoder := json.NewDecoder(response.Body)
		var jsonResponse readStruct
		err = decoder.Decode(&jsonResponse)
		if err != nil {
			log.Println("error decoding json")
			return 1
		}
		versionchan <- notifyBox + " " + jsonResponse.Version + "\n"
		return 0
	} else {
		pageContent, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("error reading version page content from " + notifyBox)
			log.Printf("%s", err)
			return 1
		}
		stringPageReturn := string(pageContent)
		versionchan <- notifyBox + " " + stringPageReturn + "\n"
		return 0
	}
}
