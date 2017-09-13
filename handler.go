package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// http handler for "/build"  begins the building process of the local managed app
func buildHandler(w http.ResponseWriter, r *http.Request) {
	gituser := r.URL.Query().Get("user")
	reponame := r.URL.Query().Get("repo")
	gobuilder := r.URL.Query().Get("gobuild")

	log.Println("checking if repo exists")
	repoResponse := repoCheck(reponame)
	if repoResponse != true {
		log.Println("repo does not exist, running git clone")
		cloneResponse := gitclone(gituser, reponame)
		if cloneResponse > 0 {
			log.Println("an error occured cloning repo")
			log.Println("user:" + gituser + ", repo:" + reponame)
			io.WriteString(w, "error cloning repo\n")
			return
		}
		log.Println("clone sucessful")
	}
	log.Println("running git pull")
	gitresponse := gitpull(gituser, reponame)
	if gitresponse > 0 {
		log.Println("an error occured running git pull")
		log.Println("user:" + gituser + ", repo:" + reponame)
		io.WriteString(w, "error running git pull\n")
		return
	}
	log.Println("git pull sucessful")
	if gobuilder == "true" {
		log.Println("running go build")
		gobuildresponse := gobuild(reponame)
		if gobuildresponse > 0 {
			log.Println("an error occured running go build")
			log.Println("reponame:" + reponame)
			io.WriteString(w, "error running go build\n")
			return
		}
		log.Println("go build sucessful")
		log.Println("copying binary")
		err := copyBinary(reponame)
		if err != nil {
			log.Println("error occured copying binary\n")
			log.Println(err)
			io.WriteString(w, "an error occured copying binary\n")
			return
		}
		log.Println("copying superviosr config")
		err = copySupervisorConf(reponame)
		if err != nil {
			log.Println("error occured copying supervisor conf\n")
			log.Println(err)
			io.WriteString(w, "error copying supervisor conf\n")
			return
		}
		log.Println("supervisor conf copied sucessfully\n")
		log.Println("restarting app in supervisor\n")
		supervisorReturn := restartSupervisor(reponame)
		if supervisorReturn > 0 {
			log.Println("an error occured restarting supervisor")
			io.WriteString(w, "error restarting supervisor\n")
			return
		}
		log.Println("supervisor restarted sucessfully")
		io.WriteString(w, "completed")
		log.Println("build sucessful")
		var config = ReadConfig()
		notifyManagers := config.Managers
		if len(notifyManagers) != 0 {
			log.Println("notifying other managers")
			notify(gituser, reponame)
			io.WriteString(w, "\nother managers have been notified"+"\n")
			log.Println("other managers have been notified")
		}
	}
}

func gitpull(gituser string, reponame string) int {
	os.Chdir(myrepos + "/" + reponame)
	cmd := "git"
	args := []string{"pull"}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func gitclone(gituser string, reponame string) int {
	os.Chdir(myrepos)
	var config = ReadConfig()
	provider := config.Provider

	cmd := exec.Command("git", "clone", provider+gituser+"/"+reponame)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return 1
	}
	log.Println("Result:  " + out.String())
	return 0
}

func gobuild(reponame string) int {
	os.Chdir(myrepos + "/" + reponame)

	cmd := exec.Command("go", "build")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return 1
	}
	log.Println("Result:  " + out.String())
	return 0
}

func repoCheck(repo string) bool {
	if _, err := os.Stat(myrepos + "/" + repo); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func copyBinary(reponame string) error {
	in, err := os.Open(myrepos + "/" + reponame + "/" + reponame)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create("/usr/local/bin/" + reponame)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	err = os.Chmod("/usr/local/bin/"+reponame, 0775)
	if err != nil {
		return err
	}
	return cerr
}

func copySupervisorConf(reponame string) error {
	in, err := os.Open(myrepos + "/" + reponame + "/" + reponame + ".conf")
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create("/etc/supervisor/conf.d/" + reponame + ".conf")
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func restartSupervisor(reponame string) int {
	//	cmd := exec.Command("supervisorctl", "restart", reponame)
	log.Println("restarting for " + reponame)
	cmd := exec.Command("service", "supervisor", "restart")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return 1
	}
	log.Println("Result:  " + out.String())
	return 0
}

//notify slave boxes to run build function
func notify(gituser string, reponame string) {
	var config = ReadConfig()
	//get list of slaves
	notifyManagers := config.Managers
	//make a channel of slaves that need to be notified
	var notifychan chan string = make(chan string)
	for i := 0; i < len(notifyManagers); i++ {
		notifyBox := notifyManagers[i]
		//notify slave $notifyBox
		go notifier(notifyBox, gituser, reponame, notifychan)
	}
	//run through list of slaves, pull response from channel
	for i := 0; i < len(notifyManagers); i++ {
		boxReturn := <-notifychan
		msgSlice := strings.Split(boxReturn, ":")
		msgBox := msgSlice[0]
		msgReturn, err := strconv.Atoi(msgSlice[1])
		if err != nil {
			log.Println(err)
		}
		if msgReturn > 0 {
			log.Println("error running notifier on: " + msgBox)
			log.Println(msgBox + " returned: " + strconv.Itoa(msgReturn))
		} else {
			log.Println(msgBox + " returned sucessfully")
		}
	}
}

//notify a slave to run it's build function
func notifier(notifyBox string, gituser string, reponame string, notifychan chan string) (int, string) {
	log.Println("notifying:  " + "http://" + notifyBox + ":8080/build?user=" + gituser + "&repo=" + reponame + "&gobuild=true")
	//send http build call
	response, err := http.Get("http://" + notifyBox + ":8080/build?user=" + gituser + "&repo=" + reponame + "&gobuild=true")
	if err != nil {
		log.Println("error making http get call to " + notifyBox)
		//add error to notify channel
		notifychan <- notifyBox + ":" + "3"
		log.Printf("%s", err)
		return 3, notifyBox
	} else {
		pageContent, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("error reading page contents on: " + notifyBox)
			log.Printf("%s", err)
			//send error to notify channel
			notifychan <- notifyBox + ":" + "2"
			return 2, notifyBox
		}
		stringPageReturn := string(pageContent)
		if stringPageReturn != "completed" {
			log.Println(notifyBox + " did not return completed")
			//send error to notification channel
			notifychan <- notifyBox + ":" + "1"
			return 1, notifyBox
		}
	}
	//send 0 to notification channel
	notifychan <- notifyBox + ":" + "0"
	return 0, notifyBox
}
