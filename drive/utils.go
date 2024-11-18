package drive

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/drive/v2"
)

func getFiles(ctx context.Context, driveService *drive.Service, folder *drive.File) (files []*drive.File) {

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

	for _, file := range files {
		fmt.Println(file.Title)
	}

	return files
}
