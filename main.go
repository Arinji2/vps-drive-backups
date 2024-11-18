package main

import (
	"fmt"
	"log"
	"os"
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

	runningBackups := driveSetup.SetupCronJobs(driveService)

	if runningBackups == 0 {
		log.Fatalf("No backups found")
	}

	server.SSHIntoServer()

	select {}
}
