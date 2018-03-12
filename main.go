package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/window"
	arg "github.com/alexflint/go-arg"
)

type argSpec struct {
	ConfigFile string `arg:"-c,--config" help:"path to configuration file"`
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
	log.SetFlags(0) // Removes default timestamp flag

	args := procArgs()
	conf, err := ParseConfigFile(args.ConfigFile)
	if err != nil {
		return err
	}

	audioFileName := filepath.Join(conf.AudioDir, conf.Play.File)
	log.Printf("using " + audioFileName)

	log.Printf("playback window: %02d:%02d to %02d:%02d",
		conf.WindowStart.Hour(), conf.WindowStart.Minute(),
		conf.WindowEnd.Hour(), conf.WindowEnd.Minute())
	win := window.New(conf.WindowStart, conf.WindowEnd)

	for {
		toWindow := win.Until()
		if toWindow == time.Duration(0) {
			log.Print("starting burst")
			for count := 0; count < conf.Play.BurstRepeat; count++ {
				err := play(audioFileName)
				if err != nil {
					return err
				}
				time.Sleep(conf.Play.IntraSleep)
			}
			log.Print("sleeping")
			time.Sleep(conf.Play.InterSleep)
		} else {
			log.Printf("sleeping until next window (%s)", toWindow)
			time.Sleep(toWindow)
		}
	}
}

func play(filename string) error {
	cmd := exec.Command("play", "-q", filename)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("play failed: %v\noutput:\n%s", err, out)
	}
	return nil
}
