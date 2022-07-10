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
	"google.golang.org/api/iterator"
	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

var MaxTranscodeJobs = 20

// SubTranscodeQueue uses GCP PubSub trigger to add jobs GCS using Transcoder API
func SubTranscodeQueue(ctx context.Context, m pubsub.Message) error {

	// Unmarshal the message into a msgTranscodeReq
	var msg msgTranscodeReq
	err := json.Unmarshal(m.Data, &msg)
	if err != nil {
		return fmt.Errorf("json unmarshal: %s", err)
	}

	// Check there are job slots available for transcoding this video
	jobsCount, err := getTranscodeJobsCount(ctx)
	if err != nil {
		return fmt.Errorf("get transcode jobs count: %s", err)
	}
	log.Printf("Transcoder Jobs Count: %d/%d", jobsCount, MaxTranscodeJobs)

	// If there is Audio and there are more than 18 jobs, don't add this job yet
	if jobsCount >= MaxTranscodeJobs {
		return fmt.Errorf("too many jobs: %d of %d", jobsCount, MaxTranscodeJobs)
	}

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
			Value: "submitted",
		},
	})

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}

	return nil
}

// getTranscodeJobsCount returns the number of jobs currently in the Transcoder API
func getTranscodeJobsCount(ctx context.Context) (int, error) {

	log.Printf("Getting Transcoder Jobs Count")

	// Get iterator for all jobs
	it := transcoderClient.ListJobs(ctx, &transcoderpb.ListJobsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
	})

	// Count the number of jobs
	count := 0
	for {
		_, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("list jobs: %s", err)
		}
		count++
	}

	return count, nil
}
