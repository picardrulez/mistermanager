package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func buildHandler(w http.ResponseWriter, r *http.Request) {
	gituser := r.URL.Query().Get("user")
	reponame := r.URL.Query().Get("repo")

	log.Println("checking if repo exists")
	repoResponse := repoCheck(reponame)
	if repoResponse != true {
		log.Println("repo does not exist, running git clone")
		cloneResponse := gitclone(gituser, reponame)
		if cloneResponse > 0 {
			log.Println("an error occured cloning repo")
			log.Println("user:" + gituser + ", repo:" + reponame)
			return
		}
		log.Println("clone sucessful")
	}
	log.Println("running git pull")
	gitresponse := gitpull(gituser, reponame)
	if gitresponse > 0 {
		log.Println("an error occured running git pull")
		log.Println("user:" + gituser + ", repo:" + reponame)
		return
	}
}

func gitpull(gituser string, reponame string) int {
	cmd := "git"
	args := []string{"-C " + myrepos + "/" + reponame + " pull"}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func gitclone(gituser string, reponame string) int {
	cmd := "git"
	args := []string{"-C " + myrepos + "/" + reponame + "clone ssh://git@github.com:" + gituser + "/" + reponame}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
func repoCheck(repo string) bool {
	if _, err := os.Stat(myrepos + "/" + repo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}