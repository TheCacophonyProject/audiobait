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
	"github.com/TheCacophonyProject/go-api"
	"github.com/TheCacophonyProject/modemd/connrequester"
)

const (
	maxInternetConnectionRetries    = 3
	retryInternetConnectionInterval = time.Minute * 100

	maxAPIConnectionRetries    = 3
	retryAPIConnectionInterval = time.Second * 30

	maxScheduleRetries    = 3
	retryScheduleInterval = time.Minute * 100
)

var errTryLater = errors.New("error getting schedule, try again later")

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

	// Make sure the path to where we keep the schedule and audio files is OK.
	if err := createAudioPath(conf.AudioDir); err != nil {
		// This is a pretty fundamental error.  We can't do anything without this.
		log.Println("Can not create audio directory.")
		return err
	}
	log.Printf("Audio files directory is %s", conf.AudioDir)

	soundCard := NewSoundCardPlayer(conf.Card, conf.VolumeControl)

	for {

		// Check internet connection.
		internetConnection := false
		var connReq *connrequester.ConnectionRequester
		for i := 1; i <= maxInternetConnectionRetries; i++ {
			connReq, err = connectToInternet()
			if err != nil {
				log.Println(err)
				log.Printf("Waiting %s to try and connect to internet again.", retryInternetConnectionInterval)
				time.Sleep(retryInternetConnectionInterval)
			} else {
				// We have internet.
				internetConnection = true
				break
			}
		}
		if !internetConnection {
			log.Println("Could not connect to the internet. Will attempt to use schedule and files on disk.")
			downloader := &Downloader{
				audioDir: conf.AudioDir,
			}
			schedule, files, err := downloader.UseAudioScheduleAndFilesOnDisk()
			if err != nil {
				log.Println(err)
			} else {
				playTodaysAudioSchedule(soundCard, files, conf.AudioDir, schedule)
			}

			// Wait until tomorrow.
			playlist.WaitUntilNextDay()
			continue
		}

		// Have internet, now try and get an API Server connection.
		APIConnection := false
		var cacAPI *api.CacophonyAPI
		for i := 1; i <= maxAPIConnectionRetries; i++ {
			cacAPI, err = tryToInitiateAPI()
			if err != nil {
				log.Println(err)
				log.Printf("Waiting %s to try and connect to API server again.", retryAPIConnectionInterval)
				time.Sleep(retryAPIConnectionInterval)
			} else {
				// We have a connection to the API Server.
				APIConnection = true
				break
			}
		}
		if !APIConnection {
			log.Println("Could not connect to the API server.")
			// Wait until tomorrow.
			playlist.WaitUntilNextDay()
			continue
		}

		// Download schedule and files. Create a downloader object to do this for us.
		downloader := &Downloader{
			cr:       connReq,
			audioDir: conf.AudioDir,
			api:      cacAPI,
		}

		// Get schedule
		gotSchedule := false
		var schedule playlist.Schedule
		for i := 1; i <= maxScheduleRetries; i++ {
			schedule, err = downloader.GetTodaysSchedule()
			if err != nil {
				log.Println(err)
				log.Printf("Waiting %s to try and download schedule.", retryScheduleInterval)
				time.Sleep(retryScheduleInterval)
			} else {
				// We have a schedule
				gotSchedule = true
				break
			}
		}
		if !gotSchedule {
			log.Println("Could not obtain a schedule.")
			// Wait until tomorrow.
			playlist.WaitUntilNextDay()
			continue
		}

		// Get files in schedule. This is tried multiple times so no need to retry here.
		files, err := downloader.GetFilesForSchedule(schedule)
		if err != nil {
			log.Println(err)
		} else {
			playTodaysAudioSchedule(soundCard, files, conf.AudioDir, schedule)
		}

		// Wait until tomorrow.
		playlist.WaitUntilNextDay()

	}
}

// Play the current day's schedule of audio lures.
func playTodaysAudioSchedule(soundCard SoundCardPlayer, files map[int]string, audioDirectory string, schedule playlist.Schedule) {
	log.Printf("Playing todays audiobait schedule...")
	player := playlist.NewPlayer(soundCard, files, audioDirectory)
	player.SetRecorder(AudioBaitEventRecorder{})
	player.PlayTodaysSchedule(schedule)

}
