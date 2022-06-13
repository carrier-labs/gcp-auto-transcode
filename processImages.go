package cloudfunctiontranscode

import (
	"context"
	"log"
)

func processImage(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Image: %s", e.Name)

	moveFile(ctx, e)

	return nil
}
