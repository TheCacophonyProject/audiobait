package playlist

import (
	"log"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/window"
)

type AudioDevice interface {
	Play(audioFileName string, volume int) error
}

type TimeManager interface {
	Now() time.Time
	Wait(time.Duration)
}

type ActualTimeManager struct{}

func (t *ActualTimeManager) Now() time.Time {
	return time.Now()
}

func (t *ActualTimeManager) Wait(duration time.Duration) {
	log.Printf("%v: Waiting for %v", time.Now(), duration)
	time.Sleep(duration)
}

type SchedulePlayer struct {
	player    AudioDevice
	time      TimeManager
	allSounds map[int]string
	filesDir  string
}

func NewPlayer(audioDevice AudioDevice, allSoundsMap map[int]string, filesDirectory string) *SchedulePlayer {
	return newSchedulePlayerWithTimeManager(audioDevice, new(ActualTimeManager), allSoundsMap, filesDirectory)
}

func newSchedulePlayerWithTimeManager(audioDevice AudioDevice,
	timeManager TimeManager,
	allSoundsMap map[int]string,
	filesDirectory string) *SchedulePlayer {
	return &SchedulePlayer{player: audioDevice, time: timeManager, allSounds: allSoundsMap, filesDir: filesDirectory}
}

func WaitUntilNextDay() {
	sp := NewPlayer(nil, nil, "")
	tomorrowStart := sp.nextDayStart()
	log.Println("No more audiobait scheduled for today. Waiting until new day starts (sometime around midday)")
	sp.time.Wait(tomorrowStart.Sub(sp.time.Now()))
}

func (sp SchedulePlayer) findNextCombo(combos []Combo) int {
	soonest := 0
	soonestTime := time.Duration(24) * time.Hour

	for count := 0; count < len(combos); count++ {
		timeUntil := sp.createWindow(combos[count]).Until()

		if timeUntil < soonestTime {
			soonest = count
			soonestTime = timeUntil
		}
	}

	return soonest
}

func (sp SchedulePlayer) IsSoundPlayingDay(schedule Schedule) bool {
	if schedule.ControlNights <= 0 {
		return true
	}

	cycleLength := schedule.PlayNights + schedule.ControlNights
	todaysStart := sp.nextDayStart().Add(-24 * time.Hour)
	dayOfCycle := todaysStart.Day() % cycleLength

	return dayOfCycle < schedule.PlayNights
}

func (sp SchedulePlayer) PlayTodaysSchedule(schedule Schedule) {
	tomorrowStart := sp.nextDayStart()
	log.Printf("Next day start is :%v", tomorrowStart)
	if sp.IsSoundPlayingDay(schedule) {
		log.Println("Today is an audio bait day.  Lets see what animals we can attract...")
		sp.PlayTodaysCombos(schedule.Combos)
	} else {
		log.Println("Today is a control day and no audiobait sounds will be played.")
	}
	log.Println("No more audiobait scheduled for today. Waiting until new day starts (sometime around midday)")

	sp.time.Wait(tomorrowStart.Sub(sp.time.Now()))
}

func (sp SchedulePlayer) PlayTodaysCombos(combos []Combo) {
	tomorrowStart := sp.nextDayStart()
	numberCombos := len(combos)
	count := sp.findNextCombo(combos)

	nextComboStart := sp.time.Now().Add(sp.createWindow(combos[count]).Until())

	for nextComboStart.Before(tomorrowStart) {
		log.Println("Playing combo...")
		sp.PlayCombo(combos[count])
		count = (count + 1) % numberCombos
		nextComboStart = sp.time.Now().Add(sp.createWindow(combos[count]).Until())
		log.Printf("Next combo start %v", nextComboStart)
	}
	log.Println("Completed playing combos for today")
}

func (sp SchedulePlayer) nextDayStart() time.Time {

	tTime := sp.time.Now()

	// days start at this time
	todayChangeOverTime := time.Date(tTime.Year(), tTime.Month(), tTime.Day(), 12, 0, 0, 0, tTime.Location())

	// If it is night time then the next day starts tomorrow
	if tTime.After(todayChangeOverTime) {
		return todayChangeOverTime.Add(24 * time.Hour)
	} else {
		return todayChangeOverTime
	}
}

func (sp SchedulePlayer) PlayCombo(combo Combo) bool {
	win := sp.createWindow(combo)
	soundChooser := NewSoundChooser(sp.allSounds)

	toWindow := win.Until()
	if toWindow > time.Duration(0) {
		log.Printf("sleeping until next window (%s)", toWindow)
		sp.time.Wait(toWindow)
		sp.playSounds(combo, soundChooser)
	}

	every := time.Duration(combo.Every) * time.Second

	for true {
		nextBurstSleep := win.UntilNextInterval(every)
		if nextBurstSleep > time.Duration(-1) {
			log.Print("Sleeping until next burst")
			sp.time.Wait(nextBurstSleep)
			sp.playSounds(combo, soundChooser)
		} else {
			log.Print("Played last burst, sleeping until end of window")
			sp.time.Wait(win.UntilEnd())
			return true
		}
	}

	return true
}

func (sp SchedulePlayer) createWindow(combo Combo) *window.Window {
	win := window.New(combo.From.Time, combo.Until.Time)
	win.Now = sp.time.Now
	return win
}

func (sp SchedulePlayer) playSounds(combo Combo, chooser *SoundChooser) {
	log.Print("Starting sound burst")
	for count := 0; count < len(combo.Sounds); count++ {
		sp.time.Wait(time.Duration(combo.Waits[count]) * time.Second)
		_, soundFilename := chooser.ChooseSound(combo.Sounds[count])
		soundFilePath := filepath.Join(sp.filesDir, soundFilename)
		log.Printf("Playing sound %s", soundFilePath)
		if err := sp.player.Play(soundFilePath, combo.Volumes[count]); err != nil {
			log.Printf("Play failed: %v", err)
		}
	}
}
