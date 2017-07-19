package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

var configfile string = "/etc/mistermanager/mm.conf"

type Config struct {
	Provider      string
	User          string
	Repo          string
	Bind          string
	VersionPath   string
	VersionFormat string
	Managers      []string
}

func ReadConfig() Config {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file not found:  ", configfile)
	}
	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	return config
}
