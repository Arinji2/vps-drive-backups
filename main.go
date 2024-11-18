package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	driveSetup "vps-drive-backups/drive"
	"vps-drive-backups/server"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("Hello, world!")

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	driveService := driveSetup.GetDriveService(b)

	if os.Getenv("BACKUPS_FOLDER_NAME") == "" {
		log.Fatalf("BACKUPS_FOLDER_NAME is not set")
	}

	runningBackups, initialBackups := driveSetup.SetupCronJobs(driveService)

	if runningBackups == 0 {
		log.Fatalf("No backups found")
	}

	server.SSHIntoServer()

	if len(initialBackups) > 0 {
		fmt.Println("Running initial backups")
		for _, backup := range initialBackups {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
				backup.RunBackup(ctx)
				cancel()
			}()

		}
	}

	select {}
}
