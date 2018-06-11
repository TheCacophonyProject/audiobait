package main

import (
	"log"
	"fmt"

	"github.com/TheCacophonyProject/audiobait/schedule"
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

	log.Printf("Audio files will be saved to %s", conf.AudioDir)

	downloader := NewDownloader()
	_, err = downloader.DownloadSchedule()
	// what to do in case of error?

	soundCard := NewSoundCardPlayer(conf.Card, conf.VolumeControl)
	player := schedule.NewSchedulePlayer(soundCard, map[int]string{101: "/var/lib/audiobait/A-Tone-His_Self-1266414414.wav"})

	fmt.Println("Playing A-Tone2")

	combo := schedule.Combo{
		From: *schedule.NewTimeOfDay("12:01"),
		Every: 20,
		Until: *schedule.NewTimeOfDay("19:01"),
		Waits: []int{0},
		Volumes:[]int{12},
		Sounds: []string{"101"},
	}

	player.PlayCombo(combo)

	return nil
}
