package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TheCacophonyProject/audiobait/api"
	"github.com/TheCacophonyProject/audiobait/playlist"
)

type Downloader struct {
	api *api.CacophonyAPI
}

func NewDownloader() *Downloader {
	return &Downloader{api: nil}
}

func (dl *Downloader) initiateAPI() error {
	log.Println("Connecting with API")
	var err error
	dl.api, err = api.Open("/etc/thermal-uploader.yaml")
	return err
}

func (dl *Downloader) DownloadSchedule() (playlist.Schedule, error) {
	if err := dl.initiateAPI(); err != nil {
		log.Printf("Could not connect to api. %s", err)
	}

	log.Println("Getting schedule")
	sch, err := dl.GetSchedule()
	if err != nil {
		return playlist.Schedule{}, err
	}

	return sch, nil
	// return sch, GetFilesFromSchedule(api, sch, filepath.Join(savePath, "files"))

}

// func DownloadSchedule(savePath string) (playlist.Schedule, error){
//   log.Println("Getting schedule")

// 	sch, err := GetSchedule(api)
// 	if err != nil {
// 	  return playlist.Schedule{}, err
// 	}
// 	log.Println(sch)
// 	log.Println("Getting files")

// 	return sch, GetFilesFromSchedule(api, sch, filepath.Join(savePath, "files"))
// }

// GetFilesFromSchedule will get all files from the IDs in the schedule and save to disk.
func (dl *Downloader) GetFilesForSchedule(schedule playlist.Schedule, fileFolder string) (map[int]string, error) {
	referencedFiles := schedule.GetReferencedSounds()

	err := os.MkdirAll(fileFolder, 0755)
	if err != nil {
		return nil, err
	}

	audioLibrary := OpenLibrary(filepath.Join(fileFolder, "audiofilelibrary.txt"))

	dl.downloadAllNewFiles(audioLibrary, referencedFiles, fileFolder)

	availableFiles := dl.listAvailableFiles(audioLibrary, referencedFiles)

	return availableFiles, nil
}

func (dl *Downloader) listAvailableFiles(audioLibrary *AudioFileLibrary, referencedFiles []int) map[int]string {
	availableFiles := make(map[int]string)
	for _, fileId := range referencedFiles {
		strFileId := strconv.Itoa(fileId)
		if filename, exists := audioLibrary.GetFileNameOnDisk(strFileId); exists {
			availableFiles[fileId] = filename
		}
	}
	return availableFiles
}

func (dl *Downloader) downloadAllNewFiles(audioLibrary *AudioFileLibrary, referencedFiles []int, fileFolder string) {
	log.Println("Starting downloading audio files.")
	for _, fileId := range referencedFiles {
		strFileId := strconv.Itoa(fileId)
		if _, exists := audioLibrary.GetFileNameOnDisk(strFileId); !exists {
			log.Printf("Attempting to download file with id %s", strFileId)

			fileInfo, err := dl.api.GetFileDetails(fileId)
			if err != nil {
				log.Printf("Could not download file with id %s.   Downloading next file" + strFileId)
			} else {
				fileNameParts := strings.Split(fileInfo.File.Details.OriginalName, ".")
				fileExt := ""
				if len(fileNameParts) > 1 {
					fileExt = "." + fileNameParts[len(fileNameParts)-1]
				}
				fileNameOnDisk := fileInfo.File.Details.Name + "-" + strFileId + fileExt

				if err = dl.api.DownloadFile(fileInfo, filepath.Join(fileFolder, fileNameOnDisk)); err != nil {
					log.Printf("Could not download file with id %s.  Error is %s. Downloading next file"+strFileId, err)
				} else {
					audioLibrary.AddFile(strFileId, fileNameOnDisk)
				}
			}
		}
	}
	log.Println("Downloading audio files complete.")
}

// GetSchedule will get the audio schedule
func (dl *Downloader) GetSchedule() (playlist.Schedule, error) {
	jsonData, err := dl.api.GetSchedule()
	if err != nil {
		return playlist.Schedule{}, err
	}
	log.Println("Audio schedule downloaded from server")

	// parse schedule
	var sr scheduleResponse
	if err := json.Unmarshal(jsonData, &sr); err != nil {
		return playlist.Schedule{}, err
	}
	log.Println("Audio schedule parsed sucessfully")

	return sr.Schedule, nil
}

type scheduleResponse struct {
	Schedule playlist.Schedule
}
