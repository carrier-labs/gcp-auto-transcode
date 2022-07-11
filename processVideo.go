package cloudfunctiontranscode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strconv"
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
		Title: strings.TrimPrefix(path.Base(e.Name), path.Ext(e.Name)),
		MetaData: dbMetaData{
			OgFile:       path.Base(e.Name),
			ContentType:  e.ContentType,
			MD5:          e.getMD5(),
			Width:        probeData.FirstVideoStream().Width,
			Height:       probeData.FirstVideoStream().Height,
			VideoCodec:   probeData.FirstVideoStream().CodecName,
			BitRate:      probeData.FirstVideoStream().BitRate,
			FrameRateAvg: probeData.FirstVideoStream().AvgFrameRate,
			SizeB:        e.getSizeB(),
			SizeMB:       e.getSizeMB(),
			Length:       duration,
		},
	}
	if probeData.FirstAudioStream() != nil {
		entry.MetaData.AudioCodec = probeData.FirstAudioStream().CodecName
	}

	// create msgTranscodeVideo to publish to pubsub
	msg := &msgTranscodeReq{
		MD5:      e.getMD5(),
		GSURI:    fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name), // gs filename
		HasAudio: probeData.FirstAudioStream() != nil,         // check if audio stream exists
		Height:   probeData.FirstVideoStream().Height,         // get video height
		Width:    probeData.FirstVideoStream().Width,          // get video width
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
	if _, err = firestoreClient.Collection("video").Doc(e.getMD5()).Set(ctx, entry); err != nil {
		return fmt.Errorf("firestore set: %s", err)
	}

	return nil
}

// processTranscodedVideoFile processes a video file uploaded to GCS
func processTranscodedVideoFile(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Transcoded Video: %s", e.Name)

	// get the MD5 from the file name
	dir := path.Dir(e.Name)
	// get last part of the path
	dir = path.Base(dir)

	// get docref from firestore for this file
	doc := firestoreClient.Collection("video").Doc(dir)

	// get just the filename  e.Name
	filename := path.Base(e.Name)

	// Update doc with file
	_, err := doc.Update(ctx, []firestore.Update{
		{
			Path:  fmt.Sprintf("versions.%s", filename),
			Value: true,
		},
	})

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}
	return nil
}
