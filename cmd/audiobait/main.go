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
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	arg "github.com/alexflint/go-arg"

	"github.com/TheCacophonyProject/audiobait/v3/audiofilelibrary"
	"github.com/TheCacophonyProject/audiobait/v3/playlist"
	goconfig "github.com/TheCacophonyProject/go-config"
)

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
	rand.Seed(time.Now().UnixNano())

	args := procArgs()
	if !args.Timestamps {
		log.SetFlags(0) // Removes default timestamp flag
	}

	log.Printf("version %s", version)
	conf, err := ParseConfig(args.ConfigDir)
	if err != nil {
		return err
	}

	if err := startService(player{
		soundCard: NewSoundCardPlayer(conf.Card, conf.VolumeControl),
		soundDir:  conf.Dir,
	}); err != nil {
		return err
	}
	log.Println("started audiobait dbus servie")

	// Make sure the path to where we keep the schedule and audio files is OK.
	if err := createAudioPath(conf.Dir); err != nil {
		// This is a pretty fundamental error.  We can't do anything without this.
		return err
	}
	log.Printf("Audio files directory is %s", conf.Dir)

	// Start checking for new schedules
	dl := NewDownloader(conf.Dir)

	var playTime <-chan time.Time
	for {
		log.Print("loading schedule from disk")
		schedulePlayer, schedule, err := createPlayer(conf.Dir)
		if err != nil {
			log.Printf("error creating player: %v (will wait for schedule update)", err)
			playTime = nil
		} else if len(schedule.Combos) < 1 {
			log.Print("No schedule defined - waiting for schedule update")
			playTime = nil
		} else {
			playIn := schedulePlayer.TimeUntilNextCombo(*schedule)
			log.Printf("waiting %s for schedule to start", playIn)
			playTime = time.After(playIn)
		}

		select {
		case <-dl.Updated():
			log.Print("new schedule - reloading")
		case <-playTime:
			log.Printf("Playing todays audiobait schedule...")
			schedulePlayer.PlayTodaysSchedule(*schedule)
		}
	}
}

func createAudioPath(audioPath string) error {
	err := os.MkdirAll(audioPath, 0755)
	if err != nil {
		return fmt.Errorf("could not create audio directory: %v", err)
	}
	return nil
}

func createPlayer(audioDirectory string) (*playlist.SchedulePlayer, *playlist.Schedule, error) {
	schedule, err := playlist.LoadScheduleFromDisk(audioDirectory)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read schedule from disk: %v", err)
	}

	files, err := getScheduleFiles(audioDirectory, schedule)
	if err != nil {
		return nil, nil, fmt.Errorf("problem collating files for schedule: %v", err)
	}

	player := playlist.NewPlayer(files, audioDirectory)
	player.SetRecorder(AudioBaitEventRecorder{})

	return player, schedule, nil
}

func getScheduleFiles(audioDirectory string, schedule *playlist.Schedule) (map[int]string, error) {
	referencedFiles := schedule.GetReferencedSounds()

	audioLibrary, err := audiofilelibrary.OpenLibrary(audioDirectory)
	if err != nil {
		return nil, fmt.Errorf("error creating audio library: %v", err)
	}

	files := make(map[int]string)
	for _, fileID := range referencedFiles {
		if filename, exists := audioLibrary.GetFileNameOnDisk(fileID); exists {
			files[fileID] = filename
		} else {
			return nil, fmt.Errorf("file for %d is missing (%s)", fileID, filename)
		}
	}
	return files, nil
}
