package cloudfunctiontranscode

// dbEntry firebase database entry
type dbEntry struct {
	Name            string     `firestore:"og-name"`
	MD5             string     `firestore:"md5"`
	ContentType     string     `firestore:"content-type"` // mime type
	MetaData        dbMetaData `firestore:"metadata,omitempty"`
	TranscodeJob    string     `firestore:"transcode-job"`    // Transcode Job ID
	TranscodeStatus string     `firestore:"transcode-status"` // Transcode Job Status
}

// dbMetaData struct to hold image/video metadata
type dbMetaData struct {
	Width        int     `firestore:"width"`          // width in pixels
	Height       int     `firestore:"height"`         // height in pixels
	SizeB        int     `firestore:"size-B"`         // size in bytes
	SizeMB       float64 `firestore:"size-MB"`        // size in MB
	Length       float64 `firestore:"length"`         // Length in seconds
	VideoCodec   string  `firestore:"video-codec"`    // Video Codec
	AudioCodec   string  `firestore:"audio-codec"`    // Audio codec
	BitRate      string  `firestore:"bitrate"`        // Bitrate in bits/second
	FrameRateAvg string  `firestore:"frame-rate-avg"` // Frame rate in frames/second
}
