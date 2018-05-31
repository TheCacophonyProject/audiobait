
package schedule

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)


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
	return fmt.Sprintf("%s: Playing %s", playTime, audioFileName);
}

func createPlayer(startTime string) (*SchedulePlayer, *TestTimeManagerAndPlayer) {
	testPlayerAndTimer := new(TestTimeManagerAndPlayer)
	testPlayerAndTimer.PlayTimes = make([]string, 0, 10)
	testPlayerAndTimer.NowTime = NewTimeOfDay(startTime).Time
	scheduleplayer := NewSchedulePlayerWithTimeManager(testPlayerAndTimer, testPlayerAndTimer)
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


func TestPlayCombos(t *testing.T) {
	combos := []Combo{createCombo("12:03", "13:08", 30, "a"),
		createCombo("21:12", "02:15", 45, "b"),
		createCombo("03:12", "06:12", 60, "c")}
	testPlayerAndTimer := new(TestTimeManagerAndPlayer)
	testPlayerAndTimer.NowTime = NewTimeOfDay("12:13").Time
	scheduleplayer := NewSchedulePlayerWithTimeManager(testPlayerAndTimer, testPlayerAndTimer)


	scheduleplayer.PlaySchedule(combos)
}


func TestFindNextCombo(t *testing.T) {
	combos := []Combo{createCombo("12:03", "15:08", 30, "a"),
										createCombo("17:12", "02:15", 45, "b"),
										createCombo("03:12", "06:12", 60, "c")}
	fmt.Print(combos[findNextCombo(combos)])
}


func createCombo(timeStart, timeEnd string, everySeconds int, soundName string) Combo {
	return Combo{From: *NewTimeOfDay(timeStart),
	Every: everySeconds * 60,
	Until: *NewTimeOfDay(timeEnd),
	Waits: []int{0},
	Volumes:[]int{10},
	Sounds: []string{soundName}}
}