package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Facebook struct {
		Token  string `json:"token"`
		PageId string `json:"pageid"`
	} `json:"facebook"`
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
