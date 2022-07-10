package cloudfunctiontranscode

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

type TranscodeJobResult struct {
	Result JobResult `json:"job"`
}

type JobResult struct {
	Name          string `json:"name"`
	State         string `json:"state"`
	FailureReason string `json:"failureReason"`
}

// TranscodeVideo uses GCP PubSub trigger to add jobs GCS using Transcoder API
func SubTranscodeResult(ctx context.Context, m pubsub.Message) error {

	// Unmarshal the data into a TranscodeJobResult
	result := TranscodeJobResult{}
	err := json.Unmarshal(m.Data, &result)
	if err != nil {
		return fmt.Errorf("json unmarshal: %s", err)
	}

	// Find the corresponding document with job name
	iter := firestoreClient.Collection("video").Where("transcode-job", "==", result.Result.Name).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// Store the job status in Firestore
		doc.Ref.Update(ctx, []firestore.Update{
			{
				Path:  "transcode-status",
				Value: result.Result.State,
			},
		})
	}

	if err != nil {
		return fmt.Errorf("firebase update: %s", err)
	}

	return nil
}
