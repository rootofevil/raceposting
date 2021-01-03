package main

import (
	"log"
	"path"

	fb "github.com/huandu/facebook"
)

func fbPublishMessage(message string) {
	entryPoint := path.Join(pageId, "feed")
	// fmt.Println(entryPoint)
	res, _ := fb.Post(entryPoint, fb.Params{
		"message":      message,
		"access_token": access_token,
		"place":        "100991768620146",
	})
	log.Printf("%+v\n", res)
}

func fbPublishPhoto(postimage string) (id string, err error) {
	file := fb.File(postimage)
	photoPoint := path.Join(pageId, "photos")
	res, err := fb.Post(photoPoint, fb.Params{
		"access_token": access_token,
		"source":       file,
	})
	if err != nil {
		log.Println(err)
		return
	}
	err = res.DecodeField("post_id", &id)
	if err != nil {
		return
	}
	return
}
