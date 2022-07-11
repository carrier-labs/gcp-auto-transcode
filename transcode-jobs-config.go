package cloudfunctiontranscode

import (
	"fmt"

	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

// jobConfigAddVideo adds builds the JobConfig for a video only job
func jobConfigVideoOnly(width int, height int) *transcoderpb.JobConfig {

	jobConfig := &transcoderpb.JobConfig{
		PubsubDestination: &transcoderpb.PubsubDestination{
			Topic: fmt.Sprintf("projects/%s/topics/%s", ProjectId, "transcode-result"),
		},
		ElementaryStreams: []*transcoderpb.ElementaryStream{
			{
				Key: "video_stream0", // Web preview Version
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
				Key: "video_stream1", // 720p Version
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
		},
	}

	// If this video is bigger than 720p then add 1080p version
	if width > 1280 || height > 720 {
		// Append ElementaryStream
		jobConfig.ElementaryStreams = append(jobConfig.ElementaryStreams, &transcoderpb.ElementaryStream{
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
		})
		// Append MuxStream
		jobConfig.MuxStreams = append(jobConfig.MuxStreams, &transcoderpb.MuxStream{
			Key:               "h265-1080p",
			Container:         "mp4",
			ElementaryStreams: []string{"video_stream2"},
		})
	}

	// If this video is bigger than 1080p then add 2160p version
	if width > 1920 || height > 1080 {
		jobConfig.ElementaryStreams = append(jobConfig.ElementaryStreams, &transcoderpb.ElementaryStream{
			Key: "video_stream3",
			ElementaryStream: &transcoderpb.ElementaryStream_VideoStream{
				VideoStream: &transcoderpb.VideoStream{
					CodecSettings: &transcoderpb.VideoStream_H265{
						H265: &transcoderpb.VideoStream_H265CodecSettings{
							BitrateBps:   35000000, // 35Mbps
							FrameRate:    60,
							HeightPixels: 2160,
							WidthPixels:  3840,
						},
					},
				},
			},
		})
		// Append MuxStream
		jobConfig.MuxStreams = append(jobConfig.MuxStreams, &transcoderpb.MuxStream{
			Key:               "h265-2160p",
			Container:         "mp4",
			ElementaryStreams: []string{"video_stream3"},
		})
	}

	return jobConfig
}

// jobConfigAddAudio adds audio to the JobConfig
func jobConfigAddAudio(config *transcoderpb.JobConfig) {

	// Append ElementaryStream for audio
	config.ElementaryStreams = append(config.ElementaryStreams, &transcoderpb.ElementaryStream{
		Key: "audio_stream0",
		ElementaryStream: &transcoderpb.ElementaryStream_AudioStream{
			AudioStream: &transcoderpb.AudioStream{
				Codec:      "aac",
				BitrateBps: 64000,
			},
		},
	})

	// Add Audio to each MuxStream
	for _, s := range config.MuxStreams {
		s.Key = fmt.Sprintf("%s-aac", s.Key)
		s.ElementaryStreams = append(s.ElementaryStreams, "audio_stream0")
	}

}
