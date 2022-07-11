// Package helloworld provides a set of Cloud Functions samples.
package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"path"

	computemd "cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	firebase "firebase.google.com/go"
)

var ProjectId, _ = computemd.ProjectID()

var storageClient *storage.Client
var firestoreClient *firestore.Client
var pubsubClient *pubsub.Client
var transcoderClient *transcoder.Client

func init() {

	// local variables
	ctx := context.Background()
	var err error

	// Create Storage Client
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Create connection to Firestore
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: ProjectId})
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Create Firestore Client
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Create PubSub Client
	pubsubClient, err = pubsub.NewClient(ctx, ProjectId)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Create Transcoder API Client
	transcoderClient, err = transcoder.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}

}

// WatchStorageBucket consumes a(ny) GCS event.
// Configure to watch google.storage.object.finalize
func WatchStorageBucket(ctx context.Context, e GCSEvent) error {

	// log out some of the fields
	log.Printf("Matching: %s", e.Name)
	log.Printf("MIME:     %s", e.ContentType)
	log.Printf("MD5Hash:  %x", e.MD5Hash)

	// Check file is in Uploads folder
	if match, _ := (path.Match("media/upload/*.*", e.Name)); match {
		log.Printf("Match: New file in 'media/upload/'")

		// Check file mime type is video or image
		switch e.getContentBaseType() {
		case "video", "image":
			if err := renameFile(ctx, e); err != nil {
				return fmt.Errorf("renameFile: %v", err)
			}
		default:
			log.Printf("Ignoring: File not video or image")
			return nil
		}
		return nil
	}

	// Match renamed video file
	if match, _ := (path.Match("media/video/*/og-*.*", e.Name)); match {
		log.Printf("Match: Renamed og-video file")
		return processOriginalVideoFile(ctx, e)
	}

	// Match renamed video file
	if match, _ := (path.Match("media/video/*/*.*", e.Name)); match {
		log.Printf("Match: Transcoded video file")
		return processTranscodedVideoFile(ctx, e)
	}

	// Match renamed image file
	if match, _ := (path.Match("media/image/*/og-*.*", e.Name)); match {
		log.Printf("Match: New renamed image file")
		return processImageFile(ctx, e)
	}

	// No Match
	log.Printf("Ignoring: File not matched")

	return nil
}
