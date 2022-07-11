package cloudfunctiontranscode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strconv"

	"cloud.google.com/go/pubsub"
)

// msgTranscodeReq holds details for requesting Transcode API Jobs on this file
type msgTranscodeReq struct {
	MD5      string `json:"md5"`       //
	FileName string `json:"file_name"` // GCS file name
	HasAudio bool   `json:"has_audio"` // has audio
	Height   int    `json:"height"`    // Height of video
	Width    int    `json:"width"`     // Width of video
}

// processVideoFile processes a video file uploaded to GCS
func processVideoFile(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Video: %s", e.Name)

	// use ffprobe to get the video's metadata
	probeData, err := probeVideoFromGCSEvent(ctx, e)
	if err != nil {
		log.Printf("ffmpeg probe error: %s", err)
	}
	log.Printf("ffmpeg probe success: %+v", probeData)

	// convert duration string to seconds float
	duration, err := strconv.ParseFloat(probeData.FirstVideoStream().Duration, 64)
	if err != nil {
		log.Printf("Error parsing duration: %s", err)
	}

	// create empty database entry
	entry := &dbEntry{
		Name:        path.Base(e.Name),
		MD5:         fmt.Sprintf("%x", e.MD5Hash),
		ContentType: e.ContentType,
		// ProbeData:   probeData,
		// MetaData: &dbMetaData{
		// 	Width:  probeData.FirstVideoStream().Width,
		// 	Height: probeData.FirstVideoStream().Height,
		// 	SizeB:  e.getSizeB(),
		// 	SizeMB: e.getSizeMB(),
		// 	Length: duration,
		// },
	}

	// create msgTranscodeVideo to publish to pubsub
	msg := &msgTranscodeReq{
		MD5:      entry.MD5,
		FileName: e.Name,                              // gs filename
		HasAudio: probeData.FirstAudioStream() != nil, // check if audio stream exists
		Height:   probeData.FirstVideoStream().Height, // get video height
		Width:    probeData.FirstVideoStream().Width,  // get video width
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
	entry.TranscodeStatus = fmt.Sprintf("Queued: %s", ServerId)

	// Log the server ID of the published message.
	log.Printf("Published message ID: %s", ServerId)

	// Log the dbEntry
	log.Printf("dbEntry: %+v", entry)

	// write new file to database
	if _, err = firestoreClient.Collection("video").Doc(entry.MD5).Set(ctx, entry); err != nil {
		return fmt.Errorf("firestore set: %s", err)
	}

	return nil
}
