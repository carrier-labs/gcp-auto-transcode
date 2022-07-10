package cloudfunctiontranscode

import (
	"fmt"

	transcoderpb "google.golang.org/genproto/googleapis/cloud/video/transcoder/v1"
)

func jobConfigVideoOnly() *transcoderpb.JobConfig {
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

func jobConfigAddAudio(config *transcoderpb.JobConfig) {

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
