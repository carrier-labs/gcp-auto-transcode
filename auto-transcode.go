// Package helloworld provides a set of Cloud Functions samples.
package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	computemd "cloud.google.com/go/compute/metadata"
)

var ProjectId, _ = computemd.ProjectID()

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Kind                    string                 `json:"kind"`
	ID                      string                 `json:"id"`
	SelfLink                string                 `json:"selfLink"`
	Name                    string                 `json:"name"`
	Bucket                  string                 `json:"bucket"`
	Generation              string                 `json:"generation"`
	Metageneration          string                 `json:"metageneration"`
	ContentType             string                 `json:"contentType"`
	TimeCreated             time.Time              `json:"timeCreated"`
	Updated                 time.Time              `json:"updated"`
	TemporaryHold           bool                   `json:"temporaryHold"`
	EventBasedHold          bool                   `json:"eventBasedHold"`
	RetentionExpirationTime time.Time              `json:"retentionExpirationTime"`
	StorageClass            string                 `json:"storageClass"`
	TimeStorageClassUpdated time.Time              `json:"timeStorageClassUpdated"`
	Size                    string                 `json:"size"`
	MD5Hash                 string                 `json:"md5Hash"`
	MediaLink               string                 `json:"mediaLink"`
	ContentEncoding         string                 `json:"contentEncoding"`
	ContentDisposition      string                 `json:"contentDisposition"`
	CacheControl            string                 `json:"cacheControl"`
	Metadata                map[string]interface{} `json:"metadata"`
	CRC32C                  string                 `json:"crc32c"`
	ComponentCount          int                    `json:"componentCount"`
	Etag                    string                 `json:"etag"`
	CustomerEncryption      struct {
		EncryptionAlgorithm string `json:"encryptionAlgorithm"`
		KeySha256           string `json:"keySha256"`
	}
	KMSKeyName    string `json:"kmsKeyName"`
	ResourceState string `json:"resourceState"`
}

// WatchStorageBucket consumes a(ny) GCS event.
// Configure to watch google.storage.object.finalize
func WatchStorageBucket(ctx context.Context, e GCSEvent) error {
	// meta, err := metadata.FromContext(ctx)
	// if err != nil {
	// 	return fmt.Errorf("metadata.FromContext: %v", err)
	// }

	// log.Printf("Event type: %v\n", meta.EventType)
	// log.Printf("Bucket: %v\n", e.Bucket)
	// log.Printf("File: %v\n", e.Name)

	gsRef := fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name)

	// Check file is in Uploads folder
	if match, _ := (path.Match("media/upload/*.*", e.Name)); !match {
		return nil
	}

	// TODO: Get type of file from video/mp4 tag

	switch getContentType(e.ContentType) {
	case "video":
		dest := fmt.Sprintf("gs://%s/media/video/%s/", e.Bucket, e.MD5Hash)
		return processVideo(gsRef, dest)
	case "image":
		return processImage(gsRef)
	}

	return nil
}

func getContentType(mime string) (Type string) {
	s := strings.Split(mime, "/")
	if len(s) > 0 {
		Type = s[0]
	}
	return
}
