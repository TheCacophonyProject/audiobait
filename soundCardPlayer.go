package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"path/filepath"
	"encoding/json"

	"github.com/godbus/dbus"
)

type SoundCardPlayer struct {
	card int
	controlName string
}

func NewSoundCardPlayer(aCard int, aControlName string) SoundCardPlayer {
  return SoundCardPlayer{card: aCard, controlName: aControlName}
}

func (p SoundCardPlayer) Play(audioFileName string, volume int) error {
	if err := p.setVolume(volume); err != nil {
		return err
	}
	if err := p.play(audioFileName); err != nil {
		return err
	}
//		queueEvent(time.Now(), audioFileName)
  return nil
}

func (p *SoundCardPlayer) setVolume(volume int) error {
	cmd := exec.Command(
		"amixer",
		"-c", fmt.Sprint(p.card),
		"sset",
		p.controlName,
		fmt.Sprintf("%d%%", volume * 10),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("volume set failed: %v\noutput:\n%s", err, out)
	}
	return nil
}

func (p *SoundCardPlayer) play(filename string) error {
	cmd := exec.Command("play", "-q", filename)
	cmd.Env = append(os.Environ(), fmt.Sprintf("AUDIODEV=hw:%d", p.card))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("play failed: %v\noutput:\n%s", err, out)
	}
	return nil
}

func (p *SoundCardPlayer) queueEvent(ts time.Time, filename string) error {
	eventDetails := map[string]interface{}{
		"description": map[string]interface{}{
			"type": "audioBait",
			"details": map[string]interface{}{
				"filename": filepath.Base(filename),
				"volume":   100,
			},
		},
	}
	detailsJSON, err := json.Marshal(&eventDetails)
	if err != nil {
		return err
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	obj := conn.Object("org.cacophony.Events", "/org/cacophony/Events")
	call := obj.Call("org.cacophony.Events.Queue", 0, detailsJSON, ts.UnixNano())
	return call.Err
}