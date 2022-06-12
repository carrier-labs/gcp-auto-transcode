package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"strings"

	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

func processVideo(uri string) error {

	log.Printf("Processing Video: %s", uri)

	// Get Transcoder API Client
	ctx := context.Background()
	c, err := transcoder.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	// Basic Transcoding
	req := &transcoderpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
		Job: &transcoderpb.Job{
			InputUri:  uri,
			OutputUri: strings.Replace(uri, "/original/", "/transcoded/", 1),
			JobConfig: &transcoderpb.Job_Config{
				Config: &transcoderpb.JobConfig{
					PubsubDestination: &transcoderpb.PubsubDestination{
						Topic: fmt.Sprintf("projects/%s/topics/%s", ProjectId, "transcode-result"),
					},
					ElementaryStreams: []*transcoderpb.ElementaryStream{
						{
							Key: "video_stream0",
							ElementaryStream: &transcoderpb.ElementaryStream_VideoStream{
								VideoStream: &transcoderpb.VideoStream{
									CodecSettings: &transcoderpb.VideoStream_H264{
										H264: &transcoderpb.VideoStream_H264CodecSettings{
											BitrateBps:   550000,
											FrameRate:    60,
											HeightPixels: 360,
											WidthPixels:  640,
										},
									},
								},
							},
						},
						{
							Key: "video_stream1",
							ElementaryStream: &transcoderpb.ElementaryStream_VideoStream{
								VideoStream: &transcoderpb.VideoStream{
									CodecSettings: &transcoderpb.VideoStream_H264{
										H264: &transcoderpb.VideoStream_H264CodecSettings{
											BitrateBps:   2500000,
											FrameRate:    60,
											HeightPixels: 720,
											WidthPixels:  1280,
										},
									},
								},
							},
						},
						{
							Key: "audio_stream0",
							ElementaryStream: &transcoderpb.ElementaryStream_AudioStream{
								AudioStream: &transcoderpb.AudioStream{
									Codec:      "aac",
									BitrateBps: 64000,
								},
							},
						},
					},
					MuxStreams: []*transcoderpb.MuxStream{
						{
							Key:               "sd",
							Container:         "mp4",
							ElementaryStreams: []string{"video_stream0", "audio_stream0"},
						},
						{
							Key:               "hd",
							Container:         "mp4",
							ElementaryStreams: []string{"video_stream1", "audio_stream0"},
						},
					},
				},
			},
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
