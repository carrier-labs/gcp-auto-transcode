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
	jobConfig := jobConfigVideoOnly()
	// If there is audio, add the audio job config
	if msg.HasAudio {
		log.Printf("Audio Present: added to job")
		jobConfigAddAudio(jobConfig)
	}

	// Request Transcoding Job (without Audio)
	resp, err := transcoderClient.CreateJob(ctx, &transcoderpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
		Job: &transcoderpb.Job{
			InputUri:  msg.FileName,
			OutputUri: strings.TrimSuffix(msg.FileName, path.Base(msg.FileName)),
			JobConfig: &transcoderpb.Job_Config{
				Config: jobConfig,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create job: %s", err)
	}

	log.Printf("Created Job: %s", resp.GetName())

	// Store the job name in Firestore
	_, err = firestoreClient.Collection("video").Doc(msg.MD5).Update(ctx, []firestore.Update{
		{
			Path:  "transcode-job",
			Value: resp.GetName(),
		},
		{
			Path:  "transcode-status",
			Value: "Processing",
		},
	})

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}

	return nil
}
