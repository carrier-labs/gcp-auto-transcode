package cloudfunctiontranscode

import (
	"context"
	"fmt"
	"log"
	"path"
)

func renameFile(ctx context.Context, e GCSEvent) error {

	// Get Storage Bucket Handle
	bucket := storageClient.Bucket(e.Bucket)

	// Get Src File Handle
	src := bucket.Object(e.Name)

	// Create dst file name
	dstFilename := fmt.Sprintf("media/%s/%x/og-%s", e.getContentBaseType(), e.MD5Hash, path.Base(e.Name))

	// Get Dst File Handle
	dst := bucket.Object(dstFilename)

	// log out
	log.Printf("Moving: %v to %v\n", src.ObjectName(), dst.ObjectName())

	// Create new file
	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("failed Object(%q).CopierFrom(%q).Run: %v", dst.ObjectName(), src.ObjectName(), err)
	}

	// Delete src file
	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf("failed Object(%q).Delete: %v", src.ObjectName(), err)
	}

	// Log move result
	log.Printf("success: Moved %v to %v\n", src.ObjectName(), dst.ObjectName())

	// Return new filename
	return nil
}
