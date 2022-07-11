package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"path"
	"strconv"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
)

// probeVideoInGCS opens a file from GCS as a stream and probes it using FFmpeg
func probeVideoInGCS(ctx context.Context, e GCSEvent) (*ffprobe.ProbeData, error) {

	// Get an io.reader from a GCS object
	r, err := storageClient.Bucket(e.Bucket).Object(e.Name).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %v", e.Name, err)
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

// getVideoMetadata uses ffprobe to get video metadata and returns it as a versionData struct
func getVideoMetadata(ctx context.Context, e GCSEvent) (versionData, error) {
	// use ffprobe to get the video's metadata
	probeData, err := probeVideoInGCS(ctx, e)
	if err != nil {
		return versionData{}, fmt.Errorf("ffmpeg probe error: %s", err)
	}
	log.Printf("ffmpeg probe success: %+v", probeData)

	// convert duration string to seconds float
	duration, err := strconv.ParseFloat(probeData.FirstVideoStream().Duration, 64)
	if err != nil {
		return versionData{}, fmt.Errorf("parsing duration: %s", err)
	}

	// Append original version to versions
	version := versionData{
		Filename:     path.Base(e.Name),
		Width:        probeData.FirstVideoStream().Width,
		Height:       probeData.FirstVideoStream().Height,
		VideoCodec:   probeData.FirstVideoStream().CodecName,
		BitRate:      probeData.FirstVideoStream().BitRate,
		FrameRateAvg: probeData.FirstVideoStream().AvgFrameRate,
		SizeB:        e.getSizeB(),
		SizeMB:       e.getSizeMB(),
		Length:       duration,
	}

	// If audio is present, append audio version to versions
	if probeData.FirstAudioStream() != nil {
		version.AudioCodec = probeData.FirstAudioStream().CodecName
	}

	return version, nil
}
