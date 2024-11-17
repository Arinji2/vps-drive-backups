package main

import (
	"fmt"
	"log"
	"os"
	driveSetup "vps-drive-backups/drive"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("Hello, world!")

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	_ = driveSetup.GetDriveService(b)

	driveSetup.SetupCronJobs()

	select {}
}
