package cloudfunctiontranscode

// dbEntry firebase database entry
type dbEntry struct {
	Title     string `firestore:"title"`
	Transcode struct {
		Ref    string `firestore:"ref"`    // Transcode ref (JobID or PubSub message)
		Status string `firestore:"status"` // Transcode Job Status
	} `firestore:"transcode"`
}

// versionData struct to hold image/video metadata
type versionData struct {
	Filename     string  `firestore:"filename"`       // Filename
	Width        int     `firestore:"width"`          // width in pixels
	Height       int     `firestore:"height"`         // height in pixels
	SizeB        int     `firestore:"size-B"`         // size in bytes
	SizeMB       float64 `firestore:"size-MB"`        // size in MB
	Length       float64 `firestore:"length"`         // Length in seconds
	VideoCodec   string  `firestore:"video-codec"`    // Video Codec
	AudioCodec   string  `firestore:"audio-codec"`    // Audio codec
	BitRate      string  `firestore:"bitrate"`        // Bitrate in bits/second
	FrameRateAvg string  `firestore:"frame-rate-avg"` // Frame rate in frames/second
	Ready        bool    `firestore:"ready"`          // Ready to transcode
}
