/*
audiobait - play sounds to lure animals for The Cacophony Project API.
Copyright (C) 2018, The Cacophony Project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TheCacophonyProject/audiobait/playlist"
	"github.com/TheCacophonyProject/go-api"
	"github.com/TheCacophonyProject/modemd/connrequester"
)

const (
	scheduleFilename  = "schedule.json"
	libraryFilename   = "audiofilelibrary.txt"
	connTimeout       = time.Minute * 2
	connRetryInterval = time.Minute * 10
	maxConnRetries    = 3
)

type Downloader struct {
	api      *api.CacophonyAPI
	audioDir string
	cr       *connrequester.ConnectionRequester
}

func NewDownloader(audioPath string) (*Downloader, error) {
	if err := createAudioPath(audioPath); err != nil {
		return nil, err
	}

	d := &Downloader{
		cr:       connrequester.NewConnectionRequester(),
		audioDir: audioPath,
	}

	log.Println("requesting internet connection")
	d.cr.Start()
	defer d.cr.Stop()
	d.cr.WaitUntilUpLoop(connTimeout, connRetryInterval, -1)
	log.Println("internet connection made")

	cacAPI, err := tryToInitiateAPI()
	if err != nil {
		return nil, err
	}
	d.api = cacAPI

	return d, nil
}

func createAudioPath(audioPath string) error {
	err := os.MkdirAll(audioPath, 0755)
	if err != nil {
		return errors.New("Could not create audioDir.  " + err.Error())
	}
	return nil
}

func tryToInitiateAPI() (*api.CacophonyAPI, error) {
	log.Println("Connecting with API")
	cacAPI, err := api.New()
	if api.IsNotRegisteredError(err) {
		log.Println("device not registered. Exiting and waiting to be restarted")
		os.Exit(0)
	} else if err != nil {
		return nil, err
	}
	return cacAPI, nil
}

func (dl *Downloader) saveScheduleToDisk(jsonData []byte) error {
	filepath := filepath.Join(dl.audioDir, scheduleFilename)
	err := ioutil.WriteFile(filepath, jsonData, 0644)
	return err
}

func (dl *Downloader) loadScheduleFromDisk() (playlist.Schedule, error) {
	filepath := filepath.Join(dl.audioDir, scheduleFilename)
	jsonData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return playlist.Schedule{}, err
	}

	var sr scheduleResponse
	if err = json.Unmarshal(jsonData, &sr); err != nil {
		return playlist.Schedule{}, err
	}

	return sr.Schedule, nil
}

func (dl *Downloader) GetTodaysSchedule() playlist.Schedule {

	if dl.api != nil {
		log.Println("Downloading schedule from server")
		dl.cr.Start()
		defer dl.cr.Stop()
		if err := dl.cr.WaitUntilUpLoop(connTimeout, connRetryInterval, maxConnRetries); err != nil {
			log.Println(err)
		} else {
			if schedule, err := dl.downloadSchedule(); err == nil {
				// success!
				return schedule
			}
			log.Printf("Failed to download schedule schedule: %s", err)
		}
	}

	// otherwise try loading from disk
	log.Println("Loading schedule from disk")
	schedule, err := dl.loadScheduleFromDisk()
	if err != nil {
		log.Printf("Failed to load schedule from disk.  %s", err)
	}

	return schedule
}

// GetFilesForSchedule will get all files from the IDs in the schedule and save to disk.
func (dl *Downloader) GetFilesForSchedule(schedule playlist.Schedule) (map[int]string, error) {
	referencedFiles := schedule.GetReferencedSounds()

	audioLibrary := OpenLibrary(filepath.Join(dl.audioDir, libraryFilename))

	dl.cr.Start()
	defer dl.cr.Stop()
	if err := dl.cr.WaitUntilUpLoop(connTimeout, connRetryInterval, maxConnRetries); err != nil {
		return nil, err
	}
	if dl.api != nil {
		dl.downloadAllNewFiles(audioLibrary, referencedFiles)
	}

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

func (dl *Downloader) downloadAllNewFiles(audioLibrary *AudioFileLibrary, referencedFiles []int) {
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

				if err = dl.api.DownloadFile(fileInfo, filepath.Join(dl.audioDir, fileNameOnDisk)); err != nil {
					log.Printf("Could not download file with id %s.  Error is %s. Downloading next file", strFileId, err)
				} else {
					audioLibrary.AddFile(strFileId, fileNameOnDisk)
				}
			}
		}
	}
	log.Println("Downloading audio files complete.")
}

// GetSchedule will get the audio schedule
func (dl *Downloader) downloadSchedule() (playlist.Schedule, error) {
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

	if err := dl.saveScheduleToDisk(jsonData); err != nil {
		log.Printf("Failed to save schedule to disk.  Error %s.", err)
	}

	return sr.Schedule, nil
}

type scheduleResponse struct {
	Schedule playlist.Schedule
}
