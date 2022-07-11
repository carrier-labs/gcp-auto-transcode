package cloudfunctiontranscode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

type TranscoderJobResult struct {
	Job struct {
		Name  string `json:"name"`
		State string `json:"state"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Details []byte `json:"details"`
		} `json:"error"`
	} `json:"job"`
}

// TranscodeVideo uses GCP PubSub trigger to add jobs GCS using Transcoder API
func SubTranscodeResult(ctx context.Context, m pubsub.Message) error {

	// Unmarshal the data into a TranscodeJobResult
	result := TranscoderJobResult{}
	err := json.Unmarshal(m.Data, &result)
	if err != nil {
		return fmt.Errorf("json unmarshal: %s", err)
	}

	// Log the PubSub message
	fmt.Printf("PubSub Message: %+v\n", m)

	// Log the job result message
	fmt.Printf("Job result: %+v\n", result)

	// Find the corresponding document with job name
	iter := firestoreClient.Collection("video").Where("transcode-job", "==", result.Job.Name).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// Found the document, log to console
		log.Printf("Found document: %s[%s]\n", doc.Ref.ID, doc.Ref.Path)

		// Store the job status in Firestore
		doc.Ref.Update(ctx, []firestore.Update{
			{
				Path:  "transcode-status",
				Value: result.Job.State,
			},
		})
	}

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}

	return nil
}
