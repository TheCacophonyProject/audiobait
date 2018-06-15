// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package playlist

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestTimeOfDay struct {
	Time TimeOfDay
}

func TestParsingValidTimeJson(t *testing.T) {
	parseJsonTimeAndCheck(t, "2:03", "02:03")
	parseJsonTimeAndCheck(t, "0:00", "00:00")
	parseJsonTimeAndCheck(t, "21:55", "21:55")
	parseJsonShouldFail(t, "25:01")
	parseJsonShouldFail(t, "20:67")

}

func parseJsonTimeAndCheck(t *testing.T, time string, checktime string) {
	timeOfDay, err := parseJsonTime(time)

	if err != nil {
		t.Errorf("Unexpected error unmarshalling: %s", err)
	}

	outputTime := fmt.Sprintf("%02d:%02d", timeOfDay.Time.Hour(), timeOfDay.Time.Minute())

	assert.Equal(t, outputTime, checktime)

	fmt.Printf("Unmarshalled as %s\n", outputTime)
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
