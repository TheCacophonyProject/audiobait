package main

import (
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

	for true {
		err = DownloadAndPlaySounds(conf.AudioDir, soundCard)
		if err != nil {
			// Wait until tomorrow.
			log.Printf("Error playing sounds %v", err)
			playlist.WaitUntilNextDay()
		}
		log.Println("Back in this loop")
	}

	return nil
}

func DownloadAndPlaySounds(audioDir string, soundCard playlist.AudioDevice) error {
	downloader := NewDownloader()
	schedule, err := downloader.DownloadSchedule()
	if err != nil {
		return err
	}

	files, err2 := downloader.GetFilesForSchedule(schedule, audioDir)
	if err2 != nil {
		return err2
	}

	log.Printf("Playing sounds")
	player := playlist.NewPlayer(soundCard, files, audioDir)
	player.PlayTodaysSchedule(schedule)
	return nil
}
