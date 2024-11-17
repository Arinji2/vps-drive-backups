package drive

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/api/drive/v2"
)

type Backup struct {
	Name       string `json:"name"`
	FolderName string `json:"folder-name"`
	BackupTime int    `json:"backup-time"`
	MaxBackups int    `json:"max-backups"`
	IsDisabled *bool  `json:"is-disabled"`
}

func SetupCronJobs(driveService *drive.Service) (runningBackups int) {

	sync.OnceFunc(func() {
		file, err := os.ReadFile("backups.json")
		if err != nil {
			log.Fatalf("Unable to read backups.json: %v", err)
		}
		var backups []Backup
		err = json.Unmarshal(file, &backups)
		if err != nil {
			log.Fatalf("Unable to unmarshal backups.json: %v", err)
		}

		for _, backup := range backups {
			fmt.Printf("Setting Up Cron Job for %s, every %d seconds\n", backup.Name, backup.BackupTime)
			if backup.IsDisabled != nil && *backup.IsDisabled {
				fmt.Printf("Skipping %s\n", backup.Name)
				continue
			}
			ticker := time.NewTicker(time.Duration(backup.BackupTime) * time.Second)
			runningBackups++
			go func(b Backup) {

				for range ticker.C {
					b.RunBackup()
				}
			}(backup)
		}

		fmt.Println("Cron Jobs Set Up")
		backupFolderName := os.Getenv("BACKUPS_FOLDER_NAME")
		created, alternateLink := verifyBackupFolder(driveService)
		if created {
			fmt.Printf("Created Backups Folder  of Name: %s It can be accessed through the URL: %s\n", backupFolderName, alternateLink)
		} else {
			fmt.Printf("Backups Folder Found of Name: %s. Folder Link: %s. \n", backupFolderName, alternateLink)
		}

		fmt.Printf("\n\n\n\n\n **FINISHED SETTING UP VPS DRIVE BACKUPS** \n\n\n\n\n")

	})()

	return runningBackups
}

// Verify backup folder exists. Creates it if not.
// Returns a bool if the folder had to be created
// Returns the alternate link of the folder
func verifyBackupFolder(driveService *drive.Service) (bool, string) {
	backupFolderName := os.Getenv("BACKUPS_FOLDER_NAME")
	folders, err := driveService.Files.List().Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and title='%s'", backupFolderName)).Do()
	if err != nil {
		log.Fatalf("Unable to list folders: %v", err)
	}
	if len(folders.Items) == 0 {
		fmt.Printf("Creating Backups Folder %s\n", backupFolderName)
		res, err := driveService.Files.Insert(&drive.File{
			MimeType: "application/vnd.google-apps.folder",
			Title:    backupFolderName,
		}).Do()

		if err != nil {
			log.Fatalf("Unable to create folder: %v", err)
		}

		return true, res.AlternateLink
	} else {
		return false, folders.Items[0].AlternateLink
	}
}

func (b Backup) RunBackup() {
	fmt.Printf("Running backup for %s\n", b.Name)
}
