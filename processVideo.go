package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

func probeVideo(ctx context.Context, e GCSEvent) (map[string]interface{}, error) {
	// use FFmpeg to get details about video
	// Provide an empty string to use default FFmpeg path
	bucket := storageClient.Bucket(e.Bucket)

	// Open file for reading
	r, err := bucket.Object(e.Name).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewReader: %s", err)
	}

	// Create and open new file
	fo, err := os.Create(path.Base(e.Name))
	if err != nil {
		log.Printf("os.Create: %s", err)
	}
	// Copy file over
	size, err := io.Copy(fo, r)
	if err != nil {
		log.Printf("io.Copy: %s", err)
	}
	log.Printf("Written: %d", size)

	return fluentffmpeg.Probe(path.Base(e.Name))

}

func processVideo(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Video: %s", e.Name)

	data, err := probeVideo(ctx, e)
	if err != nil {
		log.Printf("ffmpeg probe error: %s", err)
	}
	log.Printf("ffprobe: %+v", data)

	// Move video
	ogFile, err := moveFile(ctx, e)
	if err != nil {
		return err
	}

	// Populate Firebase
	type dbEntry struct {
		Name   string  `firestore:"og-name"`
		MD5    string  `firestore:"md5"`
		Mime   string  `firestore:"mime"`
		SizeB  int     `firestore:"size-B"`
		SizeMB float64 `firestore:"size-MB"`
	}

	entry := dbEntry{
		Name:   path.Base(e.Name),
		MD5:    fmt.Sprintf("%x", e.MD5Hash),
		SizeB:  e.SizeB,
		SizeMB: e.SizeMB,
		Mime:   e.ContentType,
	}

	_, err = firestoreClient.Collection("video").Doc(entry.MD5).Set(ctx, entry)

	if err != nil {
		return err
	}

	// Get Transcoder API Client
	c, err := transcoder.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	// Request Transcoding Job (without Audio)
	resp, err := c.CreateJob(ctx, &transcoderpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
		Job: &transcoderpb.Job{
			InputUri:  ogFile,
			OutputUri: strings.TrimSuffix(ogFile, path.Base(ogFile)),
			JobConfig: &transcoderpb.Job_Config{
				Config: jobConfigWithoutAudio(),
			},
		},
	})
	if err != nil {
		return err
	}

	log.Printf("Video Transcode Job: %s", resp.GetName())

	// Request Transcoding Job (with Audio)
	resp, err = c.CreateJob(ctx, &transcoderpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", ProjectId, "europe-west4"),
		Job: &transcoderpb.Job{
			InputUri:  ogFile,
			OutputUri: strings.TrimSuffix(ogFile, path.Base(ogFile)),
			JobConfig: &transcoderpb.Job_Config{
				Config: jobConfigWithAudio(),
			},
		},
	})
	if err != nil {
		return err
	}

	log.Printf("Video+Audio Transcode Job: %s", resp.GetName())

	return nil
}

func jobConfigWithoutAudio() *transcoderpb.JobConfig {
	return &transcoderpb.JobConfig{
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
								BitrateBps:   1500000, // 1.5Mbps
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
						CodecSettings: &transcoderpb.VideoStream_H265{
							H265: &transcoderpb.VideoStream_H265CodecSettings{
								BitrateBps:   7500000, // 7.5Mbps
								FrameRate:    60,
								HeightPixels: 720,
								WidthPixels:  1280,
							},
						},
					},
				},
			},
			{
				Key: "video_stream2",
				ElementaryStream: &transcoderpb.ElementaryStream_VideoStream{
					VideoStream: &transcoderpb.VideoStream{
						CodecSettings: &transcoderpb.VideoStream_H265{
							H265: &transcoderpb.VideoStream_H265CodecSettings{
								BitrateBps:   12000000, // 12Mbps
								FrameRate:    60,
								HeightPixels: 1080,
								WidthPixels:  1920,
							},
						},
					},
				},
			},
			{
				Key: "video_stream3",
				ElementaryStream: &transcoderpb.ElementaryStream_VideoStream{
					VideoStream: &transcoderpb.VideoStream{
						CodecSettings: &transcoderpb.VideoStream_H265{
							H265: &transcoderpb.VideoStream_H265CodecSettings{
								// BitrateBps:   60000000, // 60Mbps
								BitrateBps:   35000000, // 35Mbps
								FrameRate:    60,
								HeightPixels: 2160,
								WidthPixels:  3840,
							},
						},
					},
				},
			},
		},
		MuxStreams: []*transcoderpb.MuxStream{
			{
				Key:               "h264-360p",
				Container:         "mp4",
				ElementaryStreams: []string{"video_stream0"},
			},
			{
				Key:               "h265-720p",
				Container:         "mp4",
				ElementaryStreams: []string{"video_stream1"},
			},
			{
				Key:               "h265-1080p",
				Container:         "mp4",
				ElementaryStreams: []string{"video_stream2"},
			},
			{
				Key:               "h265-2160p",
				Container:         "mp4",
				ElementaryStreams: []string{"video_stream3"},
			},
		},
	}
}

func jobConfigWithAudio() *transcoderpb.JobConfig {

	config := jobConfigWithoutAudio()

	config.ElementaryStreams = append(config.ElementaryStreams, &transcoderpb.ElementaryStream{
		Key: "audio_stream0",
		ElementaryStream: &transcoderpb.ElementaryStream_AudioStream{
			AudioStream: &transcoderpb.AudioStream{
				Codec:      "aac",
				BitrateBps: 64000,
			},
		},
	})

	config.MuxStreams = []*transcoderpb.MuxStream{
		{
			Key:               "h264-360p-audio",
			Container:         "mp4",
			ElementaryStreams: []string{"video_stream0", "audio_stream0"},
		},
		{
			Key:               "h265-720p-audio",
			Container:         "mp4",
			ElementaryStreams: []string{"video_stream1", "audio_stream0"},
		},
		{
			Key:               "h265-1080p-audio",
			Container:         "mp4",
			ElementaryStreams: []string{"video_stream2", "audio_stream0"},
		},
		{
			Key:               "h265-2160p-audio",
			Container:         "mp4",
			ElementaryStreams: []string{"video_stream3", "audio_stream0"},
		},
	}

	return config
}
