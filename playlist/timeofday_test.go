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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestTimeOfDay struct {
	Time TimeOfDay
}

func TestParsingValidTimeJson(t *testing.T) {
	parseJsonTimeAndCheck(t, newTime(2, 3), "02:03")
	parseJsonTimeAndCheck(t, newTime(2, 3), "2:03")
	parseJsonTimeAndCheck(t, newTime(0, 0), "00:00")
	parseJsonTimeAndCheck(t, newTime(21, 55), "21:55")
	parseJsonShouldFail(t, "25:01")
	parseJsonShouldFail(t, "20:67")
}

func newTime(hour, minute int) time.Time {
	return time.Date(0, 1, 1, hour, minute, 0, 0, &time.Location{})
}

func parseJsonTimeAndCheck(t *testing.T, timeExpected time.Time, timeStr string) {
	timeOfDay := NewTimeOfDay(timeStr)
	assert.Equal(t, timeExpected, timeOfDay.Time)
	marshaledTime, err := json.Marshal(timeOfDay)
	assert.NoError(t, err)
	var newTimeOfDay TimeOfDay
	assert.NoError(t, json.Unmarshal(marshaledTime, &newTimeOfDay))
	assert.Equal(t, timeOfDay, &newTimeOfDay, "should be the same after marshaled and unmarshaled")
	assert.Equal(t, timeExpected, newTimeOfDay.Time)
}

func parseJsonShouldFail(t *testing.T, time string) {
	if _, err := parseJsonTime(time); err == nil {
		t.Errorf("Should not have parsed time correctly: %s", time)
	} else {
		fmt.Println("Invalid time didn't parse (this is the expected result).")
	}
}

func parseJsonTime(time string) (TestTimeOfDay, error) {
	var timeOfDay TestTimeOfDay
	fmt.Printf("Parsing time '%s'.\n", time)

	data := []byte(fmt.Sprintf(`{"time": "%s"}`, time))

	err := json.Unmarshal(data, &timeOfDay)

	return timeOfDay, err
}
