package schedule

import "time"

type TimeOfDay struct {
  time.Time
}

const timeLayoutJson = `"15:04"`
const timeLayout = `15:04`

func (timeOfDay *TimeOfDay) UnmarshalJSON(bValue []byte) (err error) {
	sValue := string(bValue)
	if sValue == "null" {
		timeOfDay.Time = time.Time{}
		return
	}
	timeOfDay.Time, err = time.Parse(timeLayoutJson, sValue)
	return
}

func NewTimeOfDay(timeOfDayString string) *TimeOfDay {
	timeOfDay := new(TimeOfDay)
	var err error
	if timeOfDay.Time, err = time.Parse(timeLayout, timeOfDayString); err != nil {
		timeOfDay.Time = time.Time{}
	}
	return timeOfDay
}