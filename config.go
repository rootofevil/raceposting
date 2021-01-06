package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Inputdir   string `json:"inputdir"`
	Outdir     string `json:"outdir"`
	Fontsdir   string `json:"fontsdir"`
	Archivedir string `json:"archivedir"`
	Fontname   string `json:"fontname"`
	Contentdir string `json:"contentdir"`
	Imagefile  string `json:"imagefile"`
}

func loadConf(path string) Config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
