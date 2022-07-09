package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
)

// ProbeVideoInGCS opens a file from GCS as a stream and probes it using FFmpeg
func ProbeVideoInGCS(ctx context.Context, bucket, name string) (*ffprobe.ProbeData, error) {

	// Get an io.reader from a GCS object
	r, err := storageClient.Bucket(bucket).Object(name).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %v", name, err)
	}

	// Cancel if cancelled
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Probe the video
	data, err := ffprobe.ProbeReader(ctx, r)
	if err != nil {
		log.Panicf("Error getting data: %v", err)
	}

	// Return the data
	return data, nil
}

// probeVideoFromGCSEvent probes a video from GCS using probeVideo
func probeVideoFromGCSEvent(ctx context.Context, e GCSEvent) (*ffprobe.ProbeData, error) {

	// Call probeVideo
	return ProbeVideoInGCS(ctx, e.Bucket, e.Name)

}
