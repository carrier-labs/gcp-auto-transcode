package cloudfunctiontranscode

import "gopkg.in/vansante/go-ffprobe.v2"

// dbEntry firebase database entry
type dbEntry struct {
	Name            string            `firestore:"og-name"`
	MD5             string            `firestore:"md5"`
	ContentType     string            `firestore:"content-type"` // mime type
	MetaData        dbMetaData        `firestore:"metadata,omitempty"`
	ProbeData       ffprobe.ProbeData `firestore:"ffprobe-result,omitempty"`
	TranscodeJob    string            `firestore:"transcode-job"`    // Transcode Job ID
	TranscodeStatus string            `firestore:"transcode-status"` // Transcode Job Status
}

// dbMetaData struct to hold image/video metadata
type dbMetaData struct {
	Width  int     `firestore:"width"`   // width in pixels
	Height int     `firestore:"height"`  // height in pixels
	SizeB  int     `firestore:"size-B"`  // size in bytes
	SizeMB float64 `firestore:"size-MB"` // size in MB
	Length float64 `firestore:"length"`  // Length in seconds
}
