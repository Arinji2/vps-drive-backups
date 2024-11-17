package drive

import (
	"context"
	"log"
	"net/http"
	"vps-drive-backups/token"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := token.TokenFromFile(tokFile)
	if err != nil {
		tok = token.GetTokenFromWeb(config)
		token.SaveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func GetDriveService(b []byte) *drive.Service {
	// Authorizes and returns a Drive client.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	return srv

}
