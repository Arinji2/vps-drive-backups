package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/drive/v2"
)

type backupConfig struct {
	Name         string `json:"name"`
	FolderName   string `json:"folder-name"`
	VPSLocation  string `json:"vps-location"`
	BackupTime   int    `json:"backup-time"`
	MaxBackups   int    `json:"max-backups"`
	IsDisabled   *bool  `json:"is-disabled"`
	RunInitially *bool  `json:"run-initially"`
}

type backupType struct {
	Config       backupConfig
	BackupMU     *sync.Mutex
	DriveService *drive.Service
}

func SetupCronJobs(driveService *drive.Service) (runningBackups int, initialBackups []backupType) {

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
			ticker := time.NewTicker(time.Duration(backup.BackupTime) * (time.Hour * 24))
			runningBackups++
			if *backup.RunInitially {
				backup := backupType{
					Config:       backup,
					BackupMU:     &sync.Mutex{},
					DriveService: driveService,
				}
				initialBackups = append(initialBackups, backup)
			}
			go func(b backupConfig) {

				backup := backupType{
					Config:       b,
					BackupMU:     &sync.Mutex{},
					DriveService: driveService,
				}

				for range ticker.C {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
					backup.RunBackup(ctx)
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

	return runningBackups, initialBackups
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

func (b backupType) RunBackup(ctx context.Context) {
	b.BackupMU.Lock()
	defer b.BackupMU.Unlock()
	defer fmt.Println("Finished backup for", b.Config.Name)
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

	files = b.verifyFiles(files)

	if len(files) == b.Config.MaxBackups {
		fmt.Printf("BACKUP: %s \n Max backups reached of %d. Deleting oldest backup", b.Config.Name, b.Config.MaxBackups)
		files = b.deleteOldBackup(ctx, files)
	}

}

func (b backupType) verifyFiles(files []*drive.File) []*drive.File {
	type backupWithDate struct {
		file        *drive.File
		backupName  string
		backupDate  time.Time
		backupIndex int
	}

	var dateBackups []backupWithDate

	for _, file := range files {

		if file.MimeType != "application/gzip" {
			continue
		}

		parts := strings.Split(file.Title, "-")
		if len(parts) < 3 {
			continue
		}

		backupName := parts[0]
		backupDate := parts[1]
		backupIndexStr := strings.Replace(parts[2], ".tar.gz", "", 1)

		if backupName != b.Config.Name {
			continue
		}

		backupIndex, err := strconv.Atoi(backupIndexStr)

		if err != nil || backupIndex > b.Config.MaxBackups {
			continue
		}

		parsedDate, err := time.Parse("02/01/06", backupDate)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}

		dateBackups = append(dateBackups, backupWithDate{
			file:        file,
			backupName:  backupName,
			backupDate:  parsedDate,
			backupIndex: backupIndex,
		})
	}

	sort.Slice(dateBackups, func(i, j int) bool {
		return dateBackups[i].backupIndex < dateBackups[j].backupIndex
	})

	var verifiedFiles []*drive.File

	for i := 0; i < len(dateBackups); i++ {

		if i > 0 && dateBackups[i].backupIndex != dateBackups[i-1].backupIndex+1 {
			continue
		}

		if i > 0 {
			timeDiff := dateBackups[i].backupDate.Sub(dateBackups[i-1].backupDate)
			expectedInterval := time.Duration(b.Config.BackupTime) * (time.Hour * 24)

			if timeDiff < expectedInterval || timeDiff > expectedInterval {
				continue
			}
		}

		verifiedFiles = append(verifiedFiles, dateBackups[i].file)
	}

	return verifiedFiles
}

func (b backupType) deleteOldBackup(ctx context.Context, files []*drive.File) []*drive.File {
	oldestBackup := files[0]
	files = files[1:]

	err := b.DriveService.Files.Delete(oldestBackup.Id).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Unable to delete oldest backup: %v", err)
	}

	var wg sync.WaitGroup
	for i, file := range files {
		wg.Add(1)

		go func(i int, file *drive.File) {
			defer wg.Done()
			parts := strings.Split(file.Title, "-")
			name := parts[0]
			date := parts[1]

			_, err := b.DriveService.Files.Update(file.Id, &drive.File{
				Title: fmt.Sprintf("%s-%s-%d.tar.gz", name, date, i+1),
			}).Context(ctx).Do()

			if err != nil {
				log.Fatalf("Unable to update backup file: %v", err)
			}

			newFileName := fmt.Sprintf("%s-%s-%d.tar.gz", name, date, i+1)
			files[i].Title = newFileName

		}(i, file)
	}

	wg.Wait()

	return files
}
