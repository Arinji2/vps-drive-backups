package drive

import (
	"testing"

	"google.golang.org/api/drive/v2"
)

func TestVerifyFiles(t *testing.T) {
	testCases := []struct {
		name          string
		inputFiles    []*drive.File
		expectedCount int
		description   string
	}{
		{
			name: "All Valid Backup Files",
			inputFiles: []*drive.File{
				{Title: "test-01/01/22-1.tar.gz", MimeType: "application/gzip"},
				{Title: "test-02/01/22-2.tar.gz", MimeType: "application/gzip"},
				{Title: "test-03/01/22-3.tar.gz", MimeType: "application/gzip"},
			},
			expectedCount: 3,
			description:   "Should include all gzip backup files with correct naming pattern",
		},
		{
			name: "Different Backup Time",
			inputFiles: []*drive.File{
				{Title: "test-01/01/22-1.tar.gz", MimeType: "application/gzip"},
				{Title: "test-02/01/22-2.tar.gz", MimeType: "application/gzip"},
				{Title: "test-04/01/22-3.tar.gz", MimeType: "application/gzip"},
			},
			expectedCount: 2,
			description:   "Should exclude backup files with backup time not matching the configured backup time",
		},
		{
			name: "Invalid Backup Format",
			inputFiles: []*drive.File{
				{Title: "test-01/01/22-1.txt", MimeType: "plain/text"},
			},
			expectedCount: 0,
			description:   "Show return 0 files if the backup file is not a gzip file",
		},
		{
			name: "Different Backup Index",
			inputFiles: []*drive.File{
				{Title: "test-01/01/22-1.tar.gz", MimeType: "application/gzip"},
				{Title: "test-02/01/22-2.tar.gz", MimeType: "application/gzip"},
				{Title: "test-03/01/22-5.tar.gz", MimeType: "application/gzip"},
			},
			expectedCount: 2,
			description:   "Should exclude backup files with backup index not in order",
		},
		{
			name: "Duplicate Backup Index",
			inputFiles: []*drive.File{
				{Title: "test-01/01/22-1.tar.gz", MimeType: "application/gzip"},
				{Title: "test-02/01/22-2.tar.gz", MimeType: "application/gzip"},
				{Title: "test-02/01/22-2.tar.gz", MimeType: "application/gzip"},
			},
			expectedCount: 2,
			description:   "Should handle 2 duplicate backup files",
		},
		{
			name: "Exceeding Max Count",
			inputFiles: []*drive.File{
				{Title: "test-01/01/22-1.tar.gz", MimeType: "application/gzip"},
				{Title: "test-02/01/22-2.tar.gz", MimeType: "application/gzip"},
				{Title: "test-03/01/22-3.tar.gz", MimeType: "application/gzip"},
				{Title: "test-04/01/22-4.tar.gz", MimeType: "application/gzip"},
			},
			expectedCount: 3,
			description:   "Should exclude backup files which exceed the max backup count",
		},
	}

	backupConfig := backupType{
		Config: backupConfig{
			Name:         "test",
			FolderName:   "test",
			BackupTime:   1,
			MaxBackups:   3,
			IsDisabled:   nil,
			RunInitially: nil,
		},
		DriveService: &drive.Service{},
	}

	for _, tc := range testCases {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			verifiedFiles := backupConfig.verifyFiles(testCase.inputFiles)

			if len(verifiedFiles) != testCase.expectedCount {
				t.Errorf("%s: Expected %d files, got %d. %s",
					testCase.name,
					testCase.expectedCount,
					len(verifiedFiles),
					testCase.description,
				)
			}
		})
	}
}
