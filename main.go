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
	"log"
	"time"

	arg "github.com/alexflint/go-arg"

	"github.com/TheCacophonyProject/audiobait/playlist"
)

const (
	maxRetries             = 4
	retryInterval          = 30 * time.Second
	updateScheduleInterval = time.Minute * 100
)

var errTryLater = errors.New("error getting schedule, try again later")
var errNoInternetConnection = errors.New("no internet connection")

// version is populated at link time via goreleaser
var version = "No version provided"

type argSpec struct {
	ConfigFile string `arg:"-c,--config" help:"path to configuration file"`
	Timestamps bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
}

func (argSpec) Version() string {
	return version
}

func procArgs() argSpec {
	var args argSpec
	args.ConfigFile = "/etc/audiobait.yaml"
	arg.MustParse(&args)
	return args
}

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	args := procArgs()
	if !args.Timestamps {
		log.SetFlags(0) // Removes default timestamp flag
	}

	log.Printf("version %s", version)
	conf, err := ParseConfigFile(args.ConfigFile)
	if err != nil {
		return err
	}

	soundCard := NewSoundCardPlayer(conf.Card, conf.VolumeControl)
	log.Printf("Audio files directory is %s", conf.AudioDir)

	for {

		soundsDownloaded := false
		for i := 1; i <= maxRetries; i++ {
			if err := DownloadAndPlaySounds(conf.AudioDir, soundCard); err == errTryLater {
				log.Println(err)
				log.Printf("waiting %s until updating schedule", updateScheduleInterval)
				time.Sleep(updateScheduleInterval)
			} else if err != nil {
				log.Println("Error dowloading sounds and schedule:", err)
				if i < maxRetries {
					log.Println("Trying again in", retryInterval)
					time.Sleep(retryInterval)
				}
			} else {
				soundsDownloaded = true
				log.Println("Successfully downloaded sounds and schedule.")
				break
			}
		}

		if !soundsDownloaded {
			log.Println("Could not download sounds and schedule. Will try again tomorrow")
		}

		// Wait until tomorrow.
		playlist.WaitUntilNextDay()
	}
}

// DownloadAndPlaySounds creates a new downloader and downloads schedules and audio files from the API server.
// In the event of there being no internet connection, the last downloaded schedule and audio files are used.
func DownloadAndPlaySounds(audioDir string, soundCard playlist.AudioDevice) error {
	for {
		downloader, err := NewDownloader(audioDir)
		if err != nil && err != errNoInternetConnection {
			return err
		}

		schedule := downloader.GetTodaysSchedule()
		if len(schedule.Combos) == 0 {
			return errTryLater
		}

		files, err := downloader.GetFilesForSchedule(schedule)
		if err != nil {
			return err
		}

		player := playlist.NewPlayer(soundCard, files, audioDir)
		player.SetRecorder(AudioBaitEventRecorder{})
		waitTime := player.TimeUntilNextCombo(schedule.Combos)

		if waitTime > updateScheduleInterval {
			log.Printf("Next combo is due to be played in %s, so will try and download schedule and sounds nearer to that time.", waitTime)
			return errTryLater
		}
		log.Printf("Playing todays audiobait schedule...")
		player.PlayTodaysSchedule(schedule)
		return nil

	}
}
