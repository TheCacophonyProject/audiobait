// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package playlist

import "time"

type TimeOfDay struct {
	time.Time
}

const timeLayout = `15:04`
const timeLayoutJson = `"` + timeLayout + `"`

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
	t, err := time.Parse(timeLayout, timeOfDayString)
	if err != nil {
		t = time.Time{}
	}
	return &TimeOfDay{Time: t}
}
