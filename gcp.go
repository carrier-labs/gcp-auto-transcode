package cloudfunctiontranscode

import (
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"
	"time"
)

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

func (e *GCSEvent) getContentBaseType() (Type string) {
	s := strings.Split(e.ContentType, "/")
	if len(s) > 0 {
		Type = s[0]
	}
	return
}

// getSizeB returns the size in bytes of the file.
func (e *GCSEvent) getSizeB() (SizeB int) {
	if e.SizeString != "" {
		SizeB, _ = strconv.Atoi(e.SizeString)
	}
	return
}

// get SizeMB returns the size in megabytes of the file rounded to 2 decimal places.
func (e *GCSEvent) getSizeMB() (SizeMB float64) {
	SizeMB = float64(e.getSizeB()) / (1 << 20)
	SizeMB = math.Round(SizeMB*100) / 100
	return
}

// getMD5 returns MD5 hash as HEx string
func (e *GCSEvent) getMD5() string {
	return fmt.Sprintf("%x", e.MD5Hash)
}

// getTitle returns the title of the file.
func (e *GCSEvent) getTitle() string {
	// Get filename from path
	s := path.Base(e.Name)
	// trim extension
	s = strings.TrimSuffix(s, path.Ext(s))
	// remove "og-" prefix
	s = strings.TrimPrefix(s, "og-")
	return s

}
