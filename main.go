// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"log"

	"github.com/TheCacophonyProject/audiobait/playlist"
	arg "github.com/alexflint/go-arg"
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
