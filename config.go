package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

//set local config file location
var configfile string = "/etc/mistermanager/mm.conf"

//build struct for reading config values
type Config struct {
	Provider      string
	User          string
	Repo          string
	Bind          string
	VersionPath   string
	VersionFormat string
	Managers      []string
}

//get config values
func ReadConfig() Config {
	//make sure config file exists
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file not found:  ", configfile)
	}
	var config Config
	//run config through toml decoder
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	return config
}
