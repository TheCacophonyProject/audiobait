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
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/audiobait/v3/audiofilelibrary"
	"github.com/TheCacophonyProject/audiobait/v3/playlist"
	"github.com/TheCacophonyProject/go-api"
	"github.com/TheCacophonyProject/modemd/connrequester"
)

const (
	// Parameters for requesting internet connectivity
	connTimeout       = time.Minute * 2
	connRetryInterval = time.Minute * 10
	maxConnRetries    = 3

	// Parameters for download attempts and
	maxDownloadRetries    = 4
	downloadRetryInterval = 30 * time.Second
)

func NewDownloader(audioDir string) *Downloader {
	dl := &Downloader{
		audioDir: audioDir,
		updated:  make(chan struct{}, 128),
		stop:     make(chan struct{}),
	}
	go dl.loop()
	return dl
}

// Downloader manages retrieving audio schedules and associated sound files from the API server.
type Downloader struct {
	audioDir string
	updated  chan struct{}
	stop     chan struct{}
}

func (dl *Downloader) Updated() <-chan struct{} {
	return dl.updated
}

func (dl *Downloader) Stop() {
	close(dl.stop)
}

func (dl *Downloader) loop() {
	// Always check for updates on starting
	nextUpdate := time.After(0)

	for {
		select {
		case <-nextUpdate:
			if changed, err := dl.update(); err != nil {
				log.Printf("schedule update failed: %v", err)
			} else if changed {
				log.Printf("schedule changed")
				dl.updated <- struct{}{}
			}
			// Randomise sleep time between 45 - 75 minutes in order to distribute load on API server
			checkSleep := time.Duration((45 + rand.Intn(30))) * time.Minute
			log.Printf("waiting for %s until next schedule check", checkSleep)
			nextUpdate = time.After(checkSleep)
		case <-dl.stop:
			return
		}
	}
}

func (dl *Downloader) update() (bool, error) {
	log.Println("requesting internet connection")
	connReq, err := connectToInternet()
	if err != nil {
		return false, err
	}
	log.Println("internet connection made")
	defer connReq.Stop()

	api, err := initiateAPI()
	if err != nil {
		return false, err
	}

	schedule, err := playlist.GetScheduleFromAPI(api)
	if err != nil {
		return false, err
	}
	log.Println("schedule downloaded")
	if err := dl.getFilesForSchedule(api, schedule); err != nil {
		return false, err
	}
	log.Println("starting downloading audio files.")
	if err := dl.getFilesForSchedule(api, schedule); err != nil {
		return false, err
	}
	log.Println("all audio files downloaded")

	return playlist.SaveScheduleIfNew(dl.audioDir, schedule)
}

func connectToInternet() (*connrequester.ConnectionRequester, error) {
	cr := connrequester.NewConnectionRequester()
	cr.Start()
	err := cr.WaitUntilUpLoop(connTimeout, connRetryInterval, maxConnRetries)
	if err != nil {
		return nil, err
	}
	return cr, nil

}

func initiateAPI() (*api.CacophonyAPI, error) {
	cacAPI, err := api.New()
	if api.IsNotRegisteredError(err) {
		log.Println("device not registered. Exiting and waiting to be restarted")
		os.Exit(0)
	}
	if err != nil {
		return nil, err
	}
	return cacAPI, nil
}

func (dl *Downloader) getFilesForSchedule(api *api.CacophonyAPI, schedule *playlist.Schedule) error {
	return dl.downloadAllNewFiles(api, schedule.GetReferencedSounds())
}

func (dl *Downloader) downloadAllNewFiles(api *api.CacophonyAPI, fileIDs []int) error {
	for _, fileID := range fileIDs {
		fileResp, err := dl.getFileDetails(api, fileID)
		if err != nil {
			return fmt.Errorf("error getting file details for file with ID %d. Error is %s", fileID, err)
		}
		if err := dl.downloadAudioFile(api, fileID, fileResp); err != nil {
			return fmt.Errorf("error downloading file %d: %v", fileID, err)
		}
	}
	return nil
}

func (dl *Downloader) getFileDetails(apiObj *api.CacophonyAPI, fileID int) (*api.FileResponse, error) {
	var fileResp *api.FileResponse
	err := retry(
		fmt.Sprintf("get details for file %d", fileID),
		func() error {
			var err error
			fileResp, err = apiObj.GetFileDetails(fileID)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	return fileResp, nil
}

// Try and download a single audio file from the API server.
func (dl *Downloader) downloadAudioFile(api *api.CacophonyAPI, fileID int, fileResp *api.FileResponse) error {
	filename := audiofilelibrary.MakeFileName(fileResp.File.Details.OriginalName, fileResp.File.Details.Name, fileID)

	return retry(
		fmt.Sprintf("download and validate file %d", fileID),
		func() error {
			// Note: DownloadFile will skip the download if the file already exists.
			if err := api.DownloadFile(fileResp, filepath.Join(dl.audioDir, filename)); err != nil {
				return err
			}
			if !dl.validateSoundFile(filepath.Join(dl.audioDir, filename), fileResp.FileSize) {
				log.Printf("%s is not valid. Removing from disk.", filename)
				if err := os.Remove(filepath.Join(dl.audioDir, filename)); err != nil {
					return fmt.Errorf("could not remove file: %v", err)
				}
				return errors.New("download was not valid")
			}
			return nil // File is valid
		},
	)
}

// Check that the sound file is valid.
func (dl *Downloader) validateSoundFile(filename string, expectedSize int) bool {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return fileInfo.Size() == int64(expectedSize)
}

func retry(label string, do func() error) error {
	log.Printf("Starting " + label)
	attempt := 0
	for {
		err := do()
		if err == nil {
			return nil
		}
		log.Printf("%s attempt failed: %v ", label, err)

		attempt++
		if attempt < maxDownloadRetries {
			log.Println("Trying again in", downloadRetryInterval)
			time.Sleep(downloadRetryInterval)
		} else {
			return fmt.Errorf("could not %s after multiple attempts", label)
		}
	}
}
