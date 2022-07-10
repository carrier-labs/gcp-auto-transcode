package main

import (
	"context"
	"log"

	"../cloudfunctiontranscode"
)

func main() {

	// Get a file feom Google Cloud Storage
	ctx := context.Background()
	var err error

	// Set filename variables
	bn := "client_1165_red-bull_signage_store"
	fn := "media/video/ffcb0850ed98bf92346e0a77971d3235/og-MI202108090215_h264_720p.mp4"

	// Request probe
	d, err := cloudfunctiontranscode.ProbeVideoInGCS(ctx, bn, fn)
	//  ProbeVideoInGCS(ctx, bn, fn)
	if err != nil {
		panic(err)
	}
	// output result to console
	log.Printf("%+v", d)
}
