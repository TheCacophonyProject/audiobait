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

	arg "github.com/alexflint/go-arg"

	"github.com/TheCacophonyProject/audiobait/playlist"
)

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
		err = DownloadAndPlaySounds(conf.AudioDir, soundCard)
		if err != nil {
			// Wait until tomorrow.
			log.Printf("Error playing sounds: %v", err)
			playlist.WaitUntilNextDay()
		}
	}
}

func DownloadAndPlaySounds(audioDir string, soundCard playlist.AudioDevice) error {
	downloader, err := NewDownloader(audioDir)
	if err != nil {
		return err
	}

	schedule := downloader.GetTodaysSchedule()
	if len(schedule.Combos) == 0 {
		return errors.New("No audio schedule for device, or no sounds to play in schedule.")
	}

	files, err := downloader.GetFilesForSchedule(schedule)
	if err != nil {
		return err
	}

	log.Printf("Playing todays audiobait schedule...")
	player := playlist.NewPlayer(soundCard, files, audioDir)
	player.SetRecorder(AudioBaitEventRecorder{})
	player.PlayTodaysSchedule(schedule)
	return nil
}
