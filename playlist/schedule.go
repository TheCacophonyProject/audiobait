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
	"strconv"
)

type Schedule struct {
	Description   string
	ControlNights int
	PlayNights    int
	StartDay      int
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
