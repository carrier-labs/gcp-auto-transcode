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
	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

var MaxTranscodeJobs = 20

// SubTranscodeQueue uses GCP PubSub trigger to add jobs GCS using Transcoder API
func SubTranscodeQueue(ctx context.Context, m pubsub.Message) error {

	log.Printf("SubTranscodeQueue: Called")

	// Unmarshal the message into a msgTranscodeReq
	var msg msgTranscodeReq
	err := json.Unmarshal(m.Data, &msg)
	if err != nil {
		return fmt.Errorf("json unmarshal: %s", err)
	}

	// Log the PubSub Data
	log.Printf("PubSub Data: %+v", msg)

	// Set job config base setup
	jobConfig := jobConfigVideoOnly(msg.Width, msg.Height)
	// If there is audio, add the audio job config
	if msg.HasAudio {
		log.Printf("Audio Present: added to job")
		jobConfigAddAudio(jobConfig)
	}

	// Request Transcoding Job (without Audio)
	resp, err := transcoderClient.CreateJob(ctx, &transcoderpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
		Job: &transcoderpb.Job{
			InputUri:  msg.GSURI,
			OutputUri: strings.TrimSuffix(msg.GSURI, path.Base(msg.GSURI)),
			JobConfig: &transcoderpb.Job_Config{
				Config: jobConfig,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create job: %s", err)
	}

	log.Printf("Created Job: %s", resp.GetName())

	// Create new firestore.Update array
	updates := []firestore.Update{
		{
			Path:  "transcode.ref",
			Value: resp.GetName(),
		},
		{
			Path:  "transcode.status",
			Value: "PROCESSING",
		},
		{
			Path:  "versions.default",
			Value: jobConfig.MuxStreams[len(jobConfig.MuxStreams)-1].FileName,
		},
	}

	// Range over the transcode config and add details to the updates
	for _, m := range jobConfig.MuxStreams {
		// Add ready field
		updates = append(updates, firestore.Update{
			Path:  fmt.Sprintf("versions.%s.ready", m.Key),
			Value: false,
		})
		// Add Filename
		updates = append(updates, firestore.Update{
			Path:  fmt.Sprintf("versions.%s.filename", m.Key),
			Value: fmt.Sprintf("%s.%s", m.Key, m.Container),
		})
		// ToDo: Add the other details to the updates
	}

	// Store the job name in Firestore
	_, err = firestoreClient.Collection("video").Doc(msg.MD5).Update(ctx, updates)

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}

	return nil
}
