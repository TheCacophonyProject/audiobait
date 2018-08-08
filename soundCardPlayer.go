// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/godbus/dbus"
)

type SoundCardPlayer struct {
	card        int
	controlName string
}

func NewSoundCardPlayer(aCard int, aControlName string) SoundCardPlayer {
	return SoundCardPlayer{card: aCard, controlName: aControlName}
}

func (p SoundCardPlayer) Play(audioFileName string, volume int) error {
	if err := p.setVolume(volume); err != nil {
		return err
	}
	return p.play(audioFileName)
}

func (p *SoundCardPlayer) setVolume(volume int) error {
	cmd := exec.Command(
		"amixer",
		"-c", fmt.Sprint(p.card),
		"sset",
		p.controlName,
		fmt.Sprintf("%d%%", volume*10),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("volume set failed: %v\noutput:\n%s", err, out)
	}
	return nil
}

func (p *SoundCardPlayer) play(filename string) error {
	cmd := exec.Command("play", "-q", filename)
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
