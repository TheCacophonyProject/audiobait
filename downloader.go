package main

import (
	"log"
	"path/filepath"
	"os"
	"strconv"
	"encoding/json"

	"github.com/TheCacophonyProject/audiobait/schedule"
	"github.com/TheCacophonyProject/audiobait/api"
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

func (dl *Downloader) DownloadSchedule() (schedule.Schedule, error) {
	if err := dl.initiateAPI(); err != nil {
		log.Println("Could not connect to api. %s", err)
	}

	log.Println("Getting schedule")
	sch, err := dl.GetSchedule()
	if err != nil {
	  return schedule.Schedule{}, err
	}

	return sch, nil
	// return sch, GetFilesFromSchedule(api, sch, filepath.Join(savePath, "files"))

}

// func DownloadSchedule(savePath string) (schedule.Schedule, error){
//   log.Println("Getting schedule")

// 	sch, err := GetSchedule(api)
// 	if err != nil {
// 	  return schedule.Schedule{}, err
// 	}
// 	log.Println(sch)
// 	log.Println("Getting files")

// 	return sch, GetFilesFromSchedule(api, sch, filepath.Join(savePath, "files"))
// }


// GetFilesFromSchedule will get all files from the IDs in the schedule and save to disk.
func GetFilesFromSchedule(api *api.CacophonyAPI, aSchedule schedule.Schedule, fileFolder string) error {
	err := os.MkdirAll(fileFolder, 0755)
	if err != nil {
		return err
	}

	for _, fileID := range aSchedule.AllSounds {
		err := api.GetFile(fileID, filepath.Join(fileFolder, strconv.Itoa(fileID)))
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSchedule will get the audio schedule
func (dl *Downloader) GetSchedule() (schedule.Schedule, error) {
	jsonData, err := dl.api.GetSchedule()
	if err != nil {
		return schedule.Schedule{}, err
	}
	log.Println("Audio schedule downloaded from server")

	// parse schedule
	var sr scheduleResponse
	if err := json.Unmarshal(jsonData, &sr); err != nil {
		return schedule.Schedule{}, err
	}
	log.Println("Audio schedule parsed sucessfully")
	log.Println(sr.Schedule)

	return sr.Schedule, nil
}

func (dl *Downloader) SaveScheduleToDisk() {

}

type scheduleResponse struct {
	Schedule schedule.Schedule
}