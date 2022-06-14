package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
)

func main() {

	// gs := "gs://client_1165_red-bull_signage_store/media/video/bCFVY1otNb/f88cAfEFICQ==/og-RedBull-BodyLanguage.mp4"

	// Create Storage Client and add to context
	storageClient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	obj := storageClient.Bucket("client_1165_red-bull_signage_store").Object("media/video/bCFVY1otNb/f88cAfEFICQ==/og-RedBull-BodyLanguage.mp4")
	a, err := obj.Attrs(context.Background())

	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%x", a.MD5)
	log.Printf("%+v", a)
}
