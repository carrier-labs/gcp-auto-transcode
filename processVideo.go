package cloudfunctiontranscode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strings"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
)

// msgTranscodeReq holds details for requesting Transcode API Jobs on this file
type msgTranscodeReq struct {
	MD5      string `json:"md5"`       //
	GSURI    string `json:"gs_uri"`    // GCS file name
	HasAudio bool   `json:"has_audio"` // has audio
	Height   int    `json:"height"`    // Height of video
	Width    int    `json:"width"`     // Width of video
}

// processOriginalVideoFile processes a video file uploaded to GCS
func processOriginalVideoFile(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Video: %s", e.Name)

	// Get video metadata
	originalVer, err := getVideoMetadata(ctx, e)
	if err != nil {
		return fmt.Errorf("getVideoMetadata: %s", err)
	}

	// create empty database entry
	entry := &dbEntry{
		Title: e.getTitle(),
	}

	// create msgTranscodeVideo to publish to pubsub
	msg := &msgTranscodeReq{
		MD5:      e.getRefMD5(),
		GSURI:    fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name), // gs filename
		HasAudio: originalVer.AudioCodec != "",                // check if audio stream exists
		Height:   originalVer.Height,                          // get video height
		Width:    originalVer.Width,                           // get video width
	}

	// convert struct to bytes
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json marshal: %s", err)
	}

	// send a new pubsub message
	ServerId, err := pubsubClient.Topic("transcode-queue").Publish(ctx, &pubsub.Message{Data: bytes}).Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub publish: %s", err)
	}

	// update the database entry with the Transcode Job ID
	entry.Transcode.Ref = ServerId
	entry.Transcode.Status = "QUEUED"

	// Log the server ID of the published message.
	log.Printf("Published message ID: %s", ServerId)

	// Log the dbEntry
	log.Printf("dbEntry: %+v", entry)

	// write new file to database
	doc := firestoreClient.Collection("video").Doc(e.getRefMD5())
	if _, err = doc.Set(ctx, entry); err != nil {
		return fmt.Errorf("firestore set: %s", err)
	}

	// Update with original version
	_, err = doc.Update(ctx, []firestore.Update{
		{
			Path:  "versions.original",
			Value: originalVer,
		},
	})
	if err != nil {
		return fmt.Errorf("firestore update: %s", err)
	}

	return nil
}

// processTranscodedVideoFile processes a video file uploaded to GCS
func processTranscodedVideoFile(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Transcoded Video: %s", e.Name)

	versionInfo, err := getVideoMetadata(ctx, e)
	if err != nil {
		return fmt.Errorf("getVideoMetadata: %s", err)
	}

	// Set ready field to true
	versionInfo.Ready = true

	// get docref from firestore for this file
	doc := firestoreClient.Collection("video").Doc(e.getRefMD5())

	// get Key from name
	key := strings.TrimSuffix(path.Base(e.Name), path.Ext(path.Base(e.Name)))

	// Update doc with file
	_, err = doc.Update(ctx, []firestore.Update{
		{
			Path:  fmt.Sprintf("versions.%s", key),
			Value: versionInfo,
		},
	})

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}
	return nil
}
