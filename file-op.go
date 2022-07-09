package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"path"
)

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
