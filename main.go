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
	goconfig "github.com/TheCacophonyProject/go-config"
)

const (
	maxRetries             = 4
	retryInterval          = 30 * time.Second
	updateScheduleInterval = time.Minute * 100
)

var errTryLater = errors.New("error getting schedule, try again later")

// version is populated at link time via goreleaser
var version = "No version provided"

type argSpec struct {
	ConfigDir  string `arg:"-c,--config" help:"path to configuration directory"`
	Timestamps bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
}

func (argSpec) Version() string {
	return version
}

func procArgs() argSpec {
	var args argSpec
	args.ConfigDir = goconfig.DefaultConfigDir
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
	conf, err := ParseConfig(args.ConfigDir)
	if err != nil {
		return err
	}

	soundCard := NewSoundCardPlayer(conf.Card, conf.VolumeControl)
	log.Printf("Audio files directory is %s", conf.Dir)

	for {

		soundsDownloaded := false
		for i := 1; i <= maxRetries; i++ {
			if err := DownloadAndPlaySounds(conf.Dir, soundCard); err == errTryLater {
				log.Println(err)
				log.Printf("waiting %s until updateing schedule", updateScheduleInterval)
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

func DownloadAndPlaySounds(audioDir string, soundCard playlist.AudioDevice) error {
	for {
		downloader, err := NewDownloader(audioDir)
		if err != nil {
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
			return errTryLater
		} else {
			log.Printf("Playing todays audiobait schedule...")
			player.PlayTodaysSchedule(schedule)
			return nil
		}
	}
}
