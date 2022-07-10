package cloudfunctiontranscode

import (
	"context"
	"log"
)

func processImageFile(ctx context.Context, e GCSEvent) error {

	log.Printf("Processing Image: %s", e.Name)
	log.Printf("ToDo: Stuff to image")

	return nil
}
