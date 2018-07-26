// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package playlist

import (
	"log"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/window"
)

// AudioDevice models the device playing the audio.
type AudioDevice interface {
	// Play plays a given audio file at a specified volume.
	Play(audioFileName string, volume int) error
}

// Clock models a clock.   That has been abstracted for unit testing.
type Clock interface {
	// Now gets the current time
	Now() time.Time
	// Wait does a synchronous wait for the given time duration.
	Wait(time.Duration)
}

// SoundPlayedRecorder gets a notification when a sound has been played.
type SoundPlayedRecorder interface {
	// OnBaitPlayed is called when the device believes audobait has been played
	OnAudioBaitPlayed(ts time.Time, fileId int, volume int)
}

// ActualClock uses the standard go time.
type ActualClock struct{}

func (t *ActualClock) Now() time.Time {
	return time.Now()
}

func (t *ActualClock) Wait(duration time.Duration) {
	log.Printf("Waiting for %v", duration)
	time.Sleep(duration)
}

// SchedulePlayer takes a schedule and a bunch of audio files and plays them at the times specified on the schedule.
type SchedulePlayer struct {
	player   AudioDevice
	time     Clock
	recorder SoundPlayedRecorder
	// allSounds is a map of audio file ID to name of audio file on disk
	allSounds map[int]string
	filesDir  string
}

// NewPlayer creates a new schedule player.
func NewPlayer(audioDevice AudioDevice, allSoundsMap map[int]string, filesDirectory string) *SchedulePlayer {
	return newSchedulePlayerWithClock(audioDevice, new(ActualClock), allSoundsMap, filesDirectory)
}

// newSchedulePlayerWithClock creates a new schedule player.  Should only be used for unit testing - use NewPlayer otherwise.
func newSchedulePlayerWithClock(audioDevice AudioDevice,
	clock Clock,
	allSoundsMap map[int]string,
	filesDirectory string) *SchedulePlayer {
	return &SchedulePlayer{player: audioDevice, time: clock, allSounds: allSoundsMap, filesDir: filesDirectory}
}

// WaitUntilNextDay calculates when out when the next audiobait day starts (typically around midday) and wait until then.
// It always uses the standard go time.
func WaitUntilNextDay() {
	log.Println("No more audiobait scheduled for today. Waiting until new day starts (sometime around midday)")
	now := time.Now()
	time.Sleep(nextDayStart(now).Sub(now))
}

// nextDayStart calculates when the next audiobait day starts (typically around midday).
func nextDayStart(now time.Time) time.Time {
	// start hour and minute today
	todayChangeOverTime := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	// If it is before now then next start of day must be 24 hours later.
	if now.After(todayChangeOverTime) {
		return todayChangeOverTime.Add(24 * time.Hour)
	} else {
		return todayChangeOverTime
	}
}

// findNextCombo takes and array of schedule combos and works out which one
// is the next one to be played (or is currently playing)
// Returns array position
func (sp SchedulePlayer) findNextCombo(combos []Combo) int {
	nextIndex := 0
	timeUntilNext := time.Duration(24) * time.Hour

	for count := 0; count < len(combos); count++ {
		timeUntil := sp.createWindow(combos[count]).Until()

		if timeUntil < timeUntilNext {
			nextIndex = count
			timeUntilNext = timeUntil
		}
	}

	return nextIndex
}

// SetRecorder sets the call back that records when a sound has successfully played
func (sp *SchedulePlayer) SetRecorder(recorder SoundPlayedRecorder) {
	sp.recorder = recorder
}

// IsSoundPlayingDay works out whether sounds should be played today.
// Having control days when we play no sound, helps to make sure that we canaccurately determine whether
// sounds are attracting more animals or not.   They may also help stop animals getting
// attuned to hearing the sounds.
func (sp SchedulePlayer) IsSoundPlayingDay(schedule Schedule) bool {
	if schedule.ControlNights <= 0 {
		return true
	}

	firstDay := schedule.StartDay
	if firstDay < 1 {
		firstDay = 1
	}

	todaysStart := sp.nextDayStart().Add(-24 * time.Hour)

	dayOfCycle := (todaysStart.Day() - firstDay) % schedule.CycleLength()
	if dayOfCycle < 0 {
		dayOfCycle += schedule.CycleLength()
	}

	return dayOfCycle < schedule.PlayNights
}

// PlayTodaysSchedule plays todays schedule or if it is a control day it waits until the start of the next day
func (sp SchedulePlayer) PlayTodaysSchedule(schedule Schedule) {
	tomorrowStart := sp.nextDayStart()
	if sp.IsSoundPlayingDay(schedule) {
		log.Println("Today is an audiobait day.  Lets see what animals we can attract...")
		sp.playTodaysCombos(schedule.Combos)
	} else {
		log.Println("Today is a control day and no audiobait sounds will be played.")
	}
	log.Println("No more audiobait scheduled for today. Waiting until new day starts (sometime around midday)")

	sp.time.Wait(tomorrowStart.Sub(sp.time.Now()))
}

// PlayTodaysCombos plays the given combos - doesn't not care whether it is a control day
func (sp SchedulePlayer) playTodaysCombos(combos []Combo) {
	tomorrowStart := sp.nextDayStart()
	numberCombos := len(combos)
	count := sp.findNextCombo(combos)

	nextComboStart := sp.time.Now().Add(sp.createWindow(combos[count]).Until())

	for nextComboStart.Before(tomorrowStart) {
		log.Println("Playing combo...")
		sp.playCombo(combos[count])
		count = (count + 1) % numberCombos
		nextComboStart = sp.time.Now().Add(sp.createWindow(combos[count]).Until())
	}
	log.Println("Completed playing combos for today")
}

// nextDayStart works out when the next playing day starts.   As the playing day starts around midday, this could actually be
// later today.
func (sp SchedulePlayer) nextDayStart() time.Time {
	return nextDayStart(sp.time.Now())
}

// playCombo plays a single combo
func (sp SchedulePlayer) playCombo(combo Combo) bool {
	const startOfIntervalFuzzyFactor = 3 * time.Second
	win := sp.createWindow(combo)
	soundChooser := NewSoundChooser(sp.allSounds)

	every := time.Duration(combo.Every)
	if every < 1 {
		every = 1
	}
	every = every * time.Second

	toWindow := win.Until()
	if win.Until() > time.Duration(0) {
		log.Printf("sleeping until next window (%s)", toWindow)
		sp.time.Wait(toWindow)
		sp.playSounds(combo, soundChooser)
	} else if win.UntilNextInterval(every) > every-startOfIntervalFuzzyFactor {
		// If we have waited we might have missed the start by milliseconds
		sp.playSounds(combo, soundChooser)
	}

	for {
		nextBurstSleep := win.UntilNextInterval(every)
		if nextBurstSleep > time.Duration(-1) {
			log.Print("Sleeping until next burst")
			sp.time.Wait(nextBurstSleep)
			sp.playSounds(combo, soundChooser)
		} else {
			log.Print("Played last burst, sleeping until near end of window")
			sp.time.Wait(win.UntilEnd()) // Stop 3s early so we don't miss the start of the next interval
			return true
		}
	}
}

// createWindow creates a window with the times specified in the combo definition
func (sp SchedulePlayer) createWindow(combo Combo) *window.Window {
	win := window.New(combo.From.Time, combo.Until.Time)
	win.Now = sp.time.Now
	return win
}

// playSounds plays the sounds for a combo.
func (sp SchedulePlayer) playSounds(combo Combo, chooser *SoundChooser) {
	log.Print("Starting sound burst")
	for count := 0; count < len(combo.Sounds); count++ {
		sp.time.Wait(time.Duration(combo.Waits[count]) * time.Second)
		file_id, soundFilename := chooser.ChooseSound(combo.Sounds[count])
		if file_id > 0 {
			soundFilePath := filepath.Join(sp.filesDir, soundFilename)
			volume := combo.Volumes[count]
			now := sp.time.Now()
			log.Printf("Playing sound %s", soundFilePath)
			if err := sp.player.Play(soundFilePath, volume); err != nil {
				log.Printf("Play failed: %v", err)
			} else if sp.recorder != nil {
				sp.recorder.OnAudioBaitPlayed(now, file_id, volume)
			}
		} else {
			log.Printf("Could not play %s.  Either sound does not exist or this option cannot be parsed.", combo.Sounds[count])
		}
	}
}
