// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package playlist

import (
	"encoding/json"
	"strconv"
)

type Schedule struct {
	Description   string
	ControlNights int
	PlayNights    int
	Combos        []Combo
	AllSounds     []int
}

type Combo struct {
	From    TimeOfDay
	Every   int
	Until   TimeOfDay
	Waits   []int
	Volumes []int
	Sounds  []string
}

func ParseJSONConfigFile(jsonAsString string, schedule *Schedule) error {
	data := []byte(jsonAsString)

	err := json.Unmarshal(data, &schedule)
	return err
}

// GetReferencedSounds finds the sound file ids that required for playing this schedule.
func (schedule *Schedule) GetReferencedSounds() []int {
	sounds := make(map[string]bool)
	for _, combo := range schedule.Combos {
		for _, sound := range combo.Sounds {
			sounds[sound] = true
		}
	}

	if sounds["random"] {
		return schedule.AllSounds
	}

	ids := make([]int, len(sounds))
	i := 0
	for key := range sounds {
		fileId, err := strconv.Atoi(key)
		if err == nil {
			ids[i] = fileId
			i++
		}
	}
	return ids[:i]
}

// CycleLength calculates how many days the play-control cycle is.
func (schedule *Schedule) CycleLength() int {
	cycle := schedule.PlayNights + schedule.ControlNights
	if cycle > 0 {
		return cycle
	}
	return 1
}
