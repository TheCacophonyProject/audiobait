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
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/godbus/dbus"
)

// SoundCardPlayer struct contains sound card player info.
type SoundCardPlayer struct {
	card        int
	controlName string
}

// NewSoundCardPlayer constructs a new sound card player variable.
func NewSoundCardPlayer(aCard int, aControlName string) SoundCardPlayer {
	return SoundCardPlayer{card: aCard, controlName: aControlName}
}

// Play plays an audio file.
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
