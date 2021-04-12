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

package playlist

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var soundFiles = map[int]string{
	1: "squeal",
	3: "beep",
	4: "tweet",
}

type TestClockAndAudioDevice struct {
	NowTime   time.Time
	PlayTimes []string
}

func (t *TestClockAndAudioDevice) Now() time.Time {
	return t.NowTime
}

func (t *TestClockAndAudioDevice) Wait(duration time.Duration) {
	t.NowTime = t.NowTime.Add(duration).Add(time.Microsecond)
}

func (t *TestClockAndAudioDevice) SetDay(day int, month time.Month) {
	now := t.NowTime
	t.NowTime = time.Date(1, month, day, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
}

func (t *TestClockAndAudioDevice) OnAudioBaitPlayed(ts time.Time, fileId int, volume int) {
	nowTimeAsString := fmt.Sprintf("%02d:%02d:%02d", ts.Hour(), ts.Minute(), ts.Second())
	playingString := registerPlaySound(nowTimeAsString, soundFiles[fileId])
	t.PlayTimes = append(t.PlayTimes, playingString)
	fmt.Println(playingString)
}

func registerPlaySound(playTime, audioFileName string) string {
	return fmt.Sprintf("%s: Playing %s", playTime, audioFileName)
}

func createPlayer(startTime string) (*SchedulePlayer, *TestClockAndAudioDevice, *fakePlayer) {
	fp := &fakePlayer{}
	fp.plays = true
	player = fp
	testPlayerAndTimer := new(TestClockAndAudioDevice)
	testPlayerAndTimer.PlayTimes = make([]string, 0, 10)
	testPlayerAndTimer.NowTime = NewTimeOfDay(startTime).Time
	scheduleplayer := newSchedulePlayerWithClock(testPlayerAndTimer, soundFiles, "")
	scheduleplayer.SetRecorder(testPlayerAndTimer)
	return scheduleplayer, testPlayerAndTimer, fp
}

type fakePlayer struct {
	plays bool
	error error
}

func (fp fakePlayer) Play(fileId, volume int) (bool, error) {
	return fp.plays, fp.error
}

func TestPlayingComboStartDuring(t *testing.T) {
	combo := createCombo("12:01", "13:03", 30, "beep")

	schedulePlayer, testRecorder, _ := createPlayer("12:13")
	schedulePlayer.playCombo(combo)

	expectedPlayTimes := []string{
		registerPlaySound("12:31:00", "beep"),
		registerPlaySound("13:01:00", "beep"),
	}

	assert.Equal(t, expectedPlayTimes, testRecorder.PlayTimes)
}

func TestPlayingComboStartBefore(t *testing.T) {
	combo := createCombo("12:01", "13:03", 30, "howl")

	schedulePlayer, testRecorder, _ := createPlayer("11:21")
	schedulePlayer.playCombo(combo)

	expectedPlayTimes := []string{
		registerPlaySound("12:01:00", "howl"),
		registerPlaySound("12:31:00", "howl"),
		registerPlaySound("13:01:00", "howl"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}

func TestPlayTodaysScheduleWithComboOverMiddayShouldPlayToEndOfComboThenStop(t *testing.T) {
	combos := []Combo{
		createCombo("19:00", "19:25", 30, "roar"),
		createCombo("11:12", "12:40", 60, "cry"),
	}

	schedulePlayer, testRecorder, _ := createPlayer("18:30")
	schedulePlayer.playTodaysCombos(combos)

	expectedPlayTimes := []string{
		registerPlaySound("19:00:00", "roar"),
		registerPlaySound("11:12:00", "cry"),
		registerPlaySound("12:12:00", "cry"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}

func TestPlayTodaysScheduleShouldLoopBackToStartOfCombosIfRequired(t *testing.T) {
	combos := []Combo{
		createCombo("03:00", "04:00", 45, "squeal"),
		createCombo("21:12", "22:00", 60, "tweet"),
	}

	schedulePlayer, testRecorder, _ := createPlayer("18:30")
	schedulePlayer.playTodaysCombos(combos)

	expectedPlayTimes := []string{
		registerPlaySound("21:12:00", "tweet"),
		registerPlaySound("03:00:00", "squeal"),
		registerPlaySound("03:45:00", "squeal"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}

func TestScheduleWithZeroControlNightsAlwaysPlays(t *testing.T) {
	schedule := Schedule{
		ControlNights: 0,
		PlayNights:    0,
		Combos:        []Combo{createCombo("21:12", "22:00", 60, "foo")},
	}
	schedulePlayer, _, _ := createPlayer("12:01")

	assert.True(t, schedulePlayer.IsSoundPlayingDay(schedule))
}

func checkPlaysOn(day int, month time.Month, t *testing.T, schedule Schedule, schedulePlayer *SchedulePlayer, clock *TestClockAndAudioDevice) {
	clock.SetDay(day, month)
	if !schedulePlayer.IsSoundPlayingDay(schedule) {
		t.Error(fmt.Sprintf("Expected sound to play on day %d of %v", day, month))
	}
}

func checkSilentOn(day int, month time.Month, t *testing.T, schedule Schedule, schedulePlayer *SchedulePlayer, clock *TestClockAndAudioDevice) {
	clock.SetDay(day, month)
	if schedulePlayer.IsSoundPlayingDay(schedule) {
		t.Error(fmt.Sprintf("Expected no sound (control day) on day %d of %v", day, month))
	}
}

func TestScheduleWithControlDaysOnlyPlaysOnPlayingDays(t *testing.T) {
	schedule := Schedule{
		ControlNights: 5,
		PlayNights:    2,
		StartDay:      3,
		Combos:        []Combo{createCombo("12:01", "20:00", 60, "foo")},
	}
	schedulePlayer, clock, _ := createPlayer("17:01")

	checkSilentOn(1, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(2, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(3, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(4, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(5, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(6, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(7, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(9, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(10, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(17, time.April, t, schedule, schedulePlayer, clock)

	schedule = Schedule{
		ControlNights: 3,
		PlayNights:    2,
		StartDay:      19,
		Combos:        []Combo{createCombo("12:01", "20:00", 60, "foo")},
	}
	schedulePlayer, clock, _ = createPlayer("17:01")

	checkSilentOn(18, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(19, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(20, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(21, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(3, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(4, time.April, t, schedule, schedulePlayer, clock)

	// check if StartDay not defined
	schedule = Schedule{
		ControlNights: 3,
		PlayNights:    2,
		Combos:        []Combo{createCombo("12:01", "20:00", 60, "foo")},
	}
	schedulePlayer, clock, _ = createPlayer("17:01")

	checkPlaysOn(1, time.April, t, schedule, schedulePlayer, clock)
	checkPlaysOn(2, time.April, t, schedule, schedulePlayer, clock)
	checkSilentOn(3, time.April, t, schedule, schedulePlayer, clock)
}

func TestPlayComboWithMultipleSoundsIncludingSame(t *testing.T) {
	combos := []Combo{createCombo("18:00", "18:55", 30, "roar")}

	// addAnotherSound(&combos[0], 1, "same")
	addAnotherSound(&combos[0], 3, "same")
	addAnotherSound(&combos[0], 2, "meow")

	schedulePlayer, testRecorder, _ := createPlayer("17:59")
	schedulePlayer.playTodaysCombos(combos)

	expectedPlayTimes := []string{
		registerPlaySound("18:00:00", "roar"),
		registerPlaySound("18:00:03", "roar"),
		registerPlaySound("18:00:05", "meow"),
		registerPlaySound("18:30:00", "roar"),
		registerPlaySound("18:30:03", "roar"),
		registerPlaySound("18:30:05", "meow"),
	}
	assert.Equal(t, expectedPlayTimes, testRecorder.PlayTimes)
}

func TestFindNextCombo(t *testing.T) {
	combos := []Combo{createCombo("12:03", "15:08", 30, "a"),
		createCombo("17:12", "02:15", 45, "b"),
		createCombo("03:12", "06:12", 60, "c")}
	schedulePlayer, _, _ := createPlayer("12:13")
	fmt.Print(combos[schedulePlayer.findNextCombo(combos)])
}

func TestRecorderIsNotCalledWhenSoundIsNotPlayed(t *testing.T) {
	combo := createCombo("12:01", "13:03", 30, "howl")

	schedulePlayer, testRecorder, testPlayer := createPlayer("12:10")
	testPlayer.plays = false
	schedulePlayer.playCombo(combo)

	expectedPlayedTimes := []string{}

	assert.Equal(t, expectedPlayedTimes, testRecorder.PlayTimes)
}

func TestRecorderIsNotCalledWhenErrorWithPlayingSound(t *testing.T) {
	combo := createCombo("12:01", "13:03", 30, "howl")

	schedulePlayer, testRecorder, testPlayer := createPlayer("12:10")
	testPlayer.error = errors.New("some error with playing audio")
	schedulePlayer.playCombo(combo)

	expectedPlayedTimes := []string{}

	assert.Equal(t, expectedPlayedTimes, testRecorder.PlayTimes)
}

func createCombo(timeStart, timeEnd string, everyMinutes int, soundName string) Combo {
	return Combo{
		From:    *NewTimeOfDay(timeStart),
		Every:   everyMinutes * 60,
		Until:   *NewTimeOfDay(timeEnd),
		Waits:   []int{0},
		Volumes: []int{10},
		Sounds:  []string{makeSoundNameForSchedule(soundName)},
	}
}

func addAnotherSound(combo *Combo, wait int, sound string) *Combo {
	combo.Waits = append(combo.Waits, wait)
	combo.Volumes = append(combo.Volumes, 400)
	combo.Sounds = append(combo.Sounds, makeSoundNameForSchedule(sound))
	return combo
}

func makeSoundNameForSchedule(soundName string) string {
	scheduleIdentifier := soundName
	if soundName != "same" && soundName != "random" {
		soundId := len(soundFiles) + 3
		soundFiles[soundId] = soundName
		scheduleIdentifier = strconv.Itoa(soundId)
	}
	return scheduleIdentifier
}
