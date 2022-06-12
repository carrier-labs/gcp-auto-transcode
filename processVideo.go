package transcode

import (
	"context"
	"fmt"

	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

func processVideo(e GCSEvent) error {

	// Get Transcoder API Client
	ctx := context.Background()
	c, err := transcoder.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	req := &transcoderpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
		Job: &transcoderpb.Job{
			InputUri: "",
		},
	}

	resp, err := c.CreateJob(ctx, req)
	if err != nil {
		return err
	}
	// TODO: Use resp.
	_ = resp

	return nil
}
