package main

import (
	"log"

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
	var sch schedule.Schedule
	sch, err = downloader.DownloadSchedule()

	if err == nil {
		files, err2 := downloader.GetFilesForSchedule(sch, conf.AudioDir)
		if err2 == nil {
			soundCard := NewSoundCardPlayer(conf.Card, conf.VolumeControl)
			player := schedule.NewSchedulePlayer(soundCard, files, conf.AudioDir)
			log.Print("Playing todays schedule")
			player.PlayTodaysSchedule(sch)
		} else {
			log.Printf("Downloaded files for schedule failed %v", err2)
		}
	} else {
		log.Printf("DownloadSchedule Failed %v", err)
	}

	return nil
}
