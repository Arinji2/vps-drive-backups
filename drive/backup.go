package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/api/drive/v2"
)

type backupConfig struct {
	Name        string `json:"name"`
	FolderName  string `json:"folder-name"`
	VPSLocation string `json:"vps-location"`
	BackupTime  int    `json:"backup-time"`
	MaxBackups  int    `json:"max-backups"`
	IsDisabled  *bool  `json:"is-disabled"`
}

type backupType struct {
	Config       backupConfig
	BackupMU     *sync.Mutex
	DriveService *drive.Service
}

func SetupCronJobs(driveService *drive.Service) (runningBackups int) {

	sync.OnceFunc(func() {
		file, err := os.ReadFile("backups.json")
		if err != nil {
			log.Fatalf("Unable to read backups.json: %v", err)
		}
		var backups []backupConfig
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
			go func(b backupConfig) {

				backup := backupType{
					Config:       b,
					BackupMU:     &sync.Mutex{},
					DriveService: driveService,
				}

				for range ticker.C {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
					backup.runBackup(ctx)
					cancel()
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

func (b backupType) runBackup(ctx context.Context) {
	b.BackupMU.Lock()
	defer b.BackupMU.Unlock()
	fmt.Printf("Running backup for %s\n", b.Config.Name)
	b.checkForOverride(ctx)

}

func (b backupType) checkForOverride(ctx context.Context) {
	fmt.Printf("Checking for override for %s\n", b.Config.Name)
	folder, err := b.DriveService.Files.List().Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and title='%s'", b.Config.FolderName)).Do()
	if err != nil {
		_, err = b.DriveService.Files.Insert(&drive.File{
			MimeType: "application/vnd.google-apps.folder",
			Title:    b.Config.FolderName,
		}).Do()

		if err != nil {
			log.Fatalf("Unable to create folder: %v. For Backup %s", err, b.Config.Name)
		}
		return
	}
	selectedFolder := folder.Items[0]

	files := getFiles(ctx, b.DriveService, selectedFolder)

	fmt.Println(files)

}
