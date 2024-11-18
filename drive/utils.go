package drive

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"google.golang.org/api/drive/v2"
)

func getFiles(ctx context.Context, driveService *drive.Service, folder *drive.File) (files []*drive.File) {

	pageToken := ""
	var rawFiles []*drive.File

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

		rawFiles = append(rawFiles, fileList.Items...)

		pageToken = fileList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	for _, file := range rawFiles {
		if verifyFileTitle(file) {
			files = append(files, file)
		}
	}

	return files
}

// Simply uses a regex to check if the file title is valid
func verifyFileTitle(file *drive.File) bool {
	pattern := `^[a-zA-Z0-9-_]+-\d{2}/\d{2}/\d{2}-\d+\.tar.gz$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(file.Title)
}
