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
	"os/exec"
	"time"

	"github.com/TheCacophonyProject/audiobait/audiofilelibrary"
	"github.com/TheCacophonyProject/event-reporter/eventclient"
)

const (
	testSound = "/var/lib/audiobait/testSound.wav"
)

type player struct {
	soundCard SoundCardPlayer
	soundDir  string
}

// Can be mocked for testing
var saveEvent = eventclient.AddEvent
var openLibrary = audiofilelibrary.OpenLibrary
var now = time.Now

func (p *player) PlayFromId(fileId, volume, priority int, event *eventclient.Event) (bool, error) {
	library, err := openLibrary(p.soundDir)
	if err != nil {
		return false, err
	}
	fileName, found := library.FilesByID[fileId]
	if !found {
		return false, fmt.Errorf("could not find file with ID %d", fileId)
	}
	log.Printf("playing '%s' at volume %d\n", fileName, volume)
	playTime := now()
	if err := p.soundCard.Play(p.soundDir+"/"+fileName, volume); err != nil {
		return false, err
	}
	if event != nil {
		if event.Type == "" {
			event.Type = "audioBait"
		}
		event.Timestamp = playTime
		if event.Details == nil {
			event.Details = map[string]interface{}{}
		}
		event.Details["fileId"] = fileId
		event.Details["volume"] = volume
		event.Details["priority"] = priority
		log.Println("saving audio event")
		log.Println(event)
		return true, saveEvent(*event)
	}
	return true, nil
}

func (p *player) PlayTestSound(volume int) error {
	return p.soundCard.Play(testSound, volume)
}

type SoundCardPlayer interface {
	Play(audioFileName string, volume int) error
}

// NewSoundCardPlayer constructs a new sound card player variable.
func NewSoundCardPlayer(aCard int, aControlName string) amixerPlayer {
	return amixerPlayer{card: aCard, controlName: aControlName}
}

// amixerPlayer is a SoundCardPlayer that uses amixer.
type amixerPlayer struct {
	card        int
	controlName string
}

// Play plays an audio file.
func (p amixerPlayer) Play(audioFileName string, volume int) error {
	if err := p.setVolume(volume); err != nil {
		return err
	}
	return p.play(audioFileName)
}

func (p *amixerPlayer) setVolume(volume int) error {
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

func (p *amixerPlayer) play(filename string) error {
	cmd := exec.Command("play", "-q", filename)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("play failed: %v\noutput:\n%s", err, out)
	}

	return nil
}
