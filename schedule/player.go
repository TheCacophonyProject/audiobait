package schedule

import (
	"time"
	"log"

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
}

func NewSchedulePlayer(audioPlayer AudioPlayer) *SchedulePlayer {
	return NewSchedulePlayerWithTimeManager(audioPlayer, new(ActualTimeManager))
}

func NewSchedulePlayerWithTimeManager(audioPlayer AudioPlayer, timeManager TimeManager) *SchedulePlayer {
  return &SchedulePlayer{ player: audioPlayer, time: timeManager}
}

func findNextCombo(combos []Combo) int {
	soonest := 0;
	soonestTime := time.Duration(24) * time.Hour

	for count := 0; count < len(combos); count++ {
		combo := combos[count]
		timeUntil := window.New(combo.From.Time, combo.Until.Time).Until()

		if (timeUntil < soonestTime) {
			soonest = count;
			soonestTime = timeUntil;
		}
	}

	return soonest
}

func (sp SchedulePlayer) PlaySchedule(combos []Combo) {
	numberCombos := len(combos)
  count := findNextCombo(combos)
	iterations := 0

	for count < numberCombos && iterations < 6 {
		sp.PlayCombo(combos[count])
		count = (count + 1) % numberCombos
		iterations++
	}
}

func (sp SchedulePlayer) PlayCombo(combo Combo) {
	win := window.New(combo.From.Time, combo.Until.Time)
	win.Now = sp.time.Now
	every := time.Duration(combo.Every) * time.Second
	finished := false

	waitTime := win.Until()
	if (waitTime > time.Duration(0)) {
		log.Printf("sleeping until next window (%s)", toWindow)
	}

	for !finished {
		// play sounds
		sp.time.Wait(waitTime)
		sp.playSounds(combo)

		waitTime := win.UntilNextInterval(every)
		if waitTime > time.Duration(-1) {
			log.Print("ended burst, sleeping until next burst")
		} else {
			log.Print("Played last burst, sleeping until end of window")
			finished = true
			sp.time.Wait(win.UntilEnd())
		}
	}
}

func (sp SchedulePlayer) playSounds(combo Combo) {
	for count := 0; count < len(combo.Sounds); count++ {
		log.Print("Starting burst")
		sp.time.Wait(time.Duration(combo.Waits[count]) * time.Second);
		sp.player.Play(combo.Sounds[count], combo.Volumes[count]);
	}
}
