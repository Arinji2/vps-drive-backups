package drive

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/api/drive/v2"
)

func getFiles(ctx context.Context, dataMu *sync.Mutex, driveService *drive.Service, folder *drive.File) (files []*drive.File) {
	dataMu.Lock()
	defer dataMu.Unlock()

	pageToken := ""

	for {
		query := fmt.Sprintf("'%s' in parents and trashed = false", folder.Id)
		fileList, err := driveService.Files.List().
			Q(query).
			Context(ctx).
			PageToken(pageToken).
			Do()

		if err != nil {
			log.Fatalf("Unable to query files: %v", err)
		}

		files = append(files, fileList.Items...)

		pageToken = fileList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return files
}
