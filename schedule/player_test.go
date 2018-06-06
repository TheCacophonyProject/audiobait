
package schedule

import (
	"fmt"
	"testing"
	"time"
	"strconv"

	"github.com/stretchr/testify/assert"
)

var soundFiles = map[int]string {
  1 : "squeal",
	3 : "beep",
	4 : "tweet",
}

type TestTimeManagerAndPlayer struct {
	NowTime time.Time
	PlayTimes []string
}

func (p *TestTimeManagerAndPlayer) Play(audioFileName string, _ int) {
	nowTimeAsString := fmt.Sprintf("%02d:%02d:%02d", p.NowTime.Hour(), p.NowTime.Minute(), p.NowTime.Second())
	playingString := registerPlaySound(nowTimeAsString, audioFileName)
	p.PlayTimes = append(p.PlayTimes, playingString)
	fmt.Println(playingString)
}

func (t *TestTimeManagerAndPlayer) Now() time.Time {
	return t.NowTime
}

func (t *TestTimeManagerAndPlayer) Wait(duration time.Duration) {
	t.NowTime = t.NowTime.Add(duration)
}

func registerPlaySound(playTime, audioFileName string) string {
	return fmt.Sprintf("%s: Playing %s", playTime, audioFileName)
}

func createPlayer(startTime string) (*SchedulePlayer, *TestTimeManagerAndPlayer) {
	testPlayerAndTimer := new(TestTimeManagerAndPlayer)
	testPlayerAndTimer.PlayTimes = make([]string, 0, 10)
	testPlayerAndTimer.NowTime = NewTimeOfDay(startTime).Time
	scheduleplayer := NewSchedulePlayerWithTimeManager(testPlayerAndTimer, testPlayerAndTimer, soundFiles)
	return scheduleplayer, testPlayerAndTimer
}

func TestPlayingComboStartDuring(t *testing.T) {
	combo := createCombo("12:01", "13:03", 30, "beep")

	schedulePlayer, testRecorder := createPlayer("12:13")
	schedulePlayer.PlayCombo(combo)

	expectedPlayTimes := []string{
		registerPlaySound("12:31:00", "beep"),
		registerPlaySound("13:01:00", "beep"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}

func TestPlayingComboStartBefore(t *testing.T) {
	combo := createCombo("12:01", "13:03", 30, "beep")

	schedulePlayer, testRecorder := createPlayer("11:21")
	schedulePlayer.PlayCombo(combo)

	expectedPlayTimes := []string{
		registerPlaySound("12:01:00", "beep"),
		registerPlaySound("12:31:00", "beep"),
		registerPlaySound("13:01:00", "beep"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}

func TestPlayTodaysScheduleWithComboOverMiddayShouldPlayToEndOfComboThenStop(t *testing.T) {
	combos := []Combo{createCombo("19:00", "19:25", 30, "roar"),
		createCombo("11:12", "12:40", 60, "beep")}

	schedulePlayer, testRecorder := createPlayer("18:30")
	schedulePlayer.PlayTodaysSchedule(combos)

	expectedPlayTimes := []string{
		registerPlaySound("19:00:00", "roar"),
		registerPlaySound("11:12:00", "beep"),
		registerPlaySound("12:12:00", "beep"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}

func TestPlayTodaysScheduleShouldLoopBackToStartOfCombosIfRequired(t *testing.T) {
	combos := []Combo{createCombo("03:00", "04:00", 45, "squeal"),
		createCombo("21:12", "22:00", 60, "tweet")}

	schedulePlayer, testRecorder := createPlayer("18:30")
	schedulePlayer.PlayTodaysSchedule(combos)

	expectedPlayTimes := []string{
		registerPlaySound("21:12:00", "tweet"),
		registerPlaySound("03:00:00", "squeal"),
		registerPlaySound("03:45:00", "squeal"),
	}

	assert.Equal(t, testRecorder.PlayTimes, expectedPlayTimes)
}


func TestPlayCombos(t *testing.T) {
	combos := []Combo{createCombo("12:03", "13:08", 30, "roar"),
		createCombo("21:12", "02:15", 45, "tweet"),
		createCombo("03:12", "06:12", 60, "squeal")}
	testPlayerAndTimer := new(TestTimeManagerAndPlayer)
	testPlayerAndTimer.NowTime = NewTimeOfDay("12:13").Time
	scheduleplayer := NewSchedulePlayerWithTimeManager(testPlayerAndTimer, testPlayerAndTimer, soundFiles)


	scheduleplayer.PlayTodaysSchedule(combos)
}


func TestFindNextCombo(t *testing.T) {
	combos := []Combo{createCombo("12:03", "15:08", 30, "a"),
										createCombo("17:12", "02:15", 45, "b"),
										createCombo("03:12", "06:12", 60, "c")}
	schedulePlayer, _ := createPlayer("12:13")
	fmt.Print(combos[schedulePlayer.findNextCombo(combos)])
}


func createCombo(timeStart, timeEnd string, everySeconds int, soundName string) Combo {
	return Combo{
		From: *NewTimeOfDay(timeStart),
		Every: everySeconds * 60,
		Until: *NewTimeOfDay(timeEnd),
		Waits: []int{0},
		Volumes:[]int{10},
		Sounds: []string{makeSoundNameForSchedule(soundName)},
  }
}

func addAnotherSound(combo *Combo, wait int, sound string) *Combo {
	combo.Waits = append(combo.Waits, wait)
	combo.Volumes	= append(combo.Volumes, 2)
	combo.Sounds = append(combo.Sounds, makeSoundNameForSchedule(sound))
	return combo
}

func makeSoundNameForSchedule(soundName string) string {
	scheduleIdentifier := soundName
	if (soundName != "same" && soundName != "random") {
		soundId := len(soundFiles) + 3
		soundFiles[soundId] = soundName
		scheduleIdentifier = strconv.Itoa(soundId)
	}
	return scheduleIdentifier;
}