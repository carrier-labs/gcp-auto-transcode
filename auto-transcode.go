// Package helloworld provides a set of Cloud Functions samples.
package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	computemd "cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/firestore"
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
	MD5Hash                 []byte                 `json:"md5Hash"`
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

var storageClient *storage.Client
var firestoreClient *firestore.Client

func init() {
	ctx := context.Background()
	var err error

	// Create Storage Client and add to context
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Open connection to Firestore
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: ProjectId})
	if err != nil {
		log.Fatalf("%s", err)
	}

	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}
}

// WatchStorageBucket consumes a(ny) GCS event.
// Configure to watch google.storage.object.finalize
func WatchStorageBucket(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing: %s", e.Name)

	// Check file is in Uploads folder
	if match, _ := (path.Match("media/upload/*.*", e.Name)); !match {
		log.Printf("Not an upload: Exit")
		return nil
	}

	log.Printf("MIME:       %s", e.ContentType)
	log.Printf("MD5Hash:    %x", e.MD5Hash)

	// Some maths on file size
	e.SizeB, _ = strconv.Atoi(e.SizeString)
	e.SizeMB = float64(e.SizeB) / (1 << 20)
	e.SizeMB = math.Round(e.SizeMB*100) / 100

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
	bucket := storageClient.Bucket(e.Bucket)

	// Get Src File
	src := bucket.Object(e.Name)

	dest := bucket.Object(fmt.Sprintf("media/%s/%x/og-%s", getContentType(e.ContentType), e.MD5Hash, path.Base(e.Name)))
	if _, err := dest.CopierFrom(src).Run(ctx); err != nil {
		return "", fmt.Errorf("Object(%q).CopierFrom(%q).Run: %v", dest.ObjectName(), src.ObjectName(), err)
	}
	if err := src.Delete(ctx); err != nil {
		return "", fmt.Errorf("Object(%q).Delete: %v", src.ObjectName(), err)
	}
	log.Printf("File %v moved to %v\n", src.ObjectName(), dest.ObjectName())

	return fmt.Sprintf("gs://%s/%s", dest.BucketName(), dest.ObjectName()), nil
}
