// Package helloworld provides a set of Cloud Functions samples.
package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	computemd "cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
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
	SizeString              string                 `json:"size"`
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
	SizeB         int
	SizeMB        float64
}

type key int

const (
	keyStorageBucket key = iota
	keyFirestoreClient
)

// WatchStorageBucket consumes a(ny) GCS event.
// Configure to watch google.storage.object.finalize
func WatchStorageBucket(ctx context.Context, e GCSEvent) error {

	// Check file is in Uploads folder
	if match, _ := (path.Match("media/upload/*.*", e.Name)); !match {
		return nil
	}

	// Some maths on file size
	e.SizeB, _ = strconv.Atoi(e.SizeString)
	e.SizeMB = float64(e.SizeB) / (1 << 20)

	// Create Storage Client and add to context
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, keyStorageBucket, storageClient.Bucket(e.Bucket))

	// Open connection to Firestore
	conf := &firebase.Config{ProjectID: ProjectId}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return err
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, keyFirestoreClient, firestoreClient)

	// TODO: Get type of file from video/mp4 tag
	switch getContentType(e.ContentType) {
	case "video":
		return processVideo(ctx, e)
	case "image":
		return processImage(ctx, e)
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

func moveFile(ctx context.Context, e GCSEvent) (string, error) {

	// Get Storage Bucket Handle
	bucket := ctx.Value(keyStorageBucket).(*storage.BucketHandle)

	// Get Src File
	src := bucket.Object(e.Name)

	dest := bucket.Object(fmt.Sprintf("media/%s/%s/og-%s", getContentType(e.ContentType), e.MD5Hash, path.Base(e.Name)))
	if _, err := dest.CopierFrom(src).Run(ctx); err != nil {
		return "", fmt.Errorf("Object(%q).CopierFrom(%q).Run: %v", dest.ObjectName(), src.ObjectName(), err)
	}
	if err := src.Delete(ctx); err != nil {
		return "", fmt.Errorf("Object(%q).Delete: %v", src.ObjectName(), err)
	}
	log.Printf("File %v moved to %v.\n", src.ObjectName(), dest.ObjectName())

	return fmt.Sprintf("gs://%s/%s", dest.BucketName(), dest.ObjectName()), nil
}
