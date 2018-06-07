package schedule

import (
	"time"
	"log"
	"fmt"

	"github.com/TheCacophonyProject/window"
)


type AudioPlayer interface {
	Play(audioFileName string, volume int)
}

type TimeManager interface {
  Now() time.Time
	Wait(time.Duration)
}

type ActualTimeManager struct {}

func (t *ActualTimeManager) Now() time.Time {
	return time.Now()
}

func (t *ActualTimeManager) Wait(duration time.Duration) {
	time.Sleep(duration)
}

type SchedulePlayer struct {
	player AudioPlayer
	time TimeManager
	allSounds map[int]string
}

func NewSchedulePlayer(audioPlayer AudioPlayer,  allSoundsMap map[int]string) *SchedulePlayer {
	return NewSchedulePlayerWithTimeManager(audioPlayer, new(ActualTimeManager), allSoundsMap)
}

func NewSchedulePlayerWithTimeManager(audioPlayer AudioPlayer, timeManager TimeManager, allSoundsMap map[int]string) *SchedulePlayer {
  return &SchedulePlayer{ player: audioPlayer, time: timeManager, allSounds: allSoundsMap}
}

func (sp SchedulePlayer) findNextCombo(combos []Combo) int {
	soonest := 0;
	soonestTime := time.Duration(24) * time.Hour

	for count := 0; count < len(combos); count++ {
		timeUntil := sp.createWindow(combos[count]).Until()

		if (timeUntil < soonestTime) {
			soonest = count;
			soonestTime = timeUntil;
		}
	}

	return soonest
}

func (sp SchedulePlayer) IsSoundPlayingDay(schedule Schedule) bool {
	if (schedule.ControlNights <= 0) {
		return true
	}

	cycleLength := schedule.PlayNights + schedule.ControlNights
	todaysStart := sp.nextDayStart().Add(-24 * time.Hour)
	dayOfCycle := todaysStart.Day() % cycleLength;

	fmt.Print("Cycle day is ")
	fmt.Print("day of Cycle")

	return dayOfCycle < schedule.PlayNights
}

func (sp SchedulePlayer) PlayTodaysSchedule(schedule Schedule) {
	if (sp.IsSoundPlayingDay(schedule)) {
		sp.PlayTodaysCombos(schedule.Combos)
	}
}

func (sp SchedulePlayer) PlayTodaysCombos(combos []Combo) {
	numberCombos := len(combos)
	tomorrowStart := sp.nextDayStart()
	count := sp.findNextCombo(combos)

	nextComboStart := sp.time.Now().Add(sp.createWindow(combos[count]).Until())

	for nextComboStart.Before(tomorrowStart) {
		sp.PlayCombo(combos[count])
		count = (count + 1) % numberCombos
		nextComboStart = sp.time.Now().Add(sp.createWindow(combos[count]).Until())
	}
}

func (sp SchedulePlayer) nextDayStart() time.Time {
	// Days start at 12 midday.
	tTime := sp.time.Now()

	// If it is night time then the next day starts tomorrow
	if (tTime.Hour() >= 12) {
		tTime = tTime.Add(14 * time.Hour)
	}

	return time.Date(tTime.Year(), tTime.Month(), tTime.Day(), 12, 00, 0, 0, tTime.Location())
}


func (sp SchedulePlayer) PlayCombo(combo Combo) bool {
	win := sp.createWindow(combo)
	soundChooser := NewSoundChooser(sp.allSounds)

	toWindow := win.Until()
	if (toWindow > time.Duration(0)) {
			log.Printf("sleeping until next window (%s)", toWindow)
			sp.time.Wait(toWindow)
			sp.playSounds(combo, soundChooser)
	}

	every := time.Duration(combo.Every) * time.Second

	for true {
		nextBurstSleep := win.UntilNextInterval(every)
		if nextBurstSleep > time.Duration(-1) {
			log.Print("ended burst, sleeping until next burst")
			sp.time.Wait(nextBurstSleep)
			sp.playSounds(combo, soundChooser)
		} else {
			log.Print("Played last burst, sleeping until end of window")
			sp.time.Wait(win.UntilEnd())
			return true
		}
	}

	return true;
}

func (sp SchedulePlayer) createWindow(combo Combo) *window.Window {
	win := window.New(combo.From.Time, combo.Until.Time)
	win.Now = sp.time.Now
	return win
}

func (sp SchedulePlayer) playSounds(combo Combo, chooser *SoundChooser) {
	for count := 0; count < len(combo.Sounds); count++ {
		log.Print("Starting burst")
		sp.time.Wait(time.Duration(combo.Waits[count]) * time.Second);
		_, soundFilename := chooser.ChooseSound(combo.Sounds[count])
		sp.player.Play(soundFilename, combo.Volumes[count]);
	}
}
