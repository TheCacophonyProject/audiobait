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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingSchedule(t *testing.T) {
	var schedule Schedule
	err := ParseJSONConfigFile(`{
		"combos": [
			{
				"from": "19:03",
				"every": 180,
				"until": "21:30",
				"waits": [
					0,
					5,
					2
				],
				"sounds": [
					"random",
					"same",
					"212"
				],
				"volumes": [
					2,
					10,
					5
				]
			},
			{
				"from": "22:00",
				"every": 270,
				"until": "02:00",
				"waits": [
					0
				],
				"sounds": [
					"215"
				],
				"volumes": [
					9
				]
			}
		],
		"playNights": 1,
		"description": "Simple schedule. ",
		"controlNights": 3,
		"allsounds": [1, 2, 3, 212, 215]
	}`, &schedule)
	if err != nil {
		t.Errorf("Error loading schedule: %s", err)
	} else {
		assert.Equal(t, schedule.Description, "Simple schedule. ")
		assert.Equal(t, schedule.ControlNights, 3)
		assert.Equal(t, len(schedule.Combos), 2)
		assert.Equal(t, schedule.Combos[0].From.Hour(), 19)
		assert.Equal(t, schedule.Combos[0].From.Minute(), 3)
		assert.Equal(t, schedule.Combos[0].Every, 180)
		assert.Equal(t, len(schedule.Combos[0].Waits), 3)
		assert.Equal(t, schedule.Combos[0].Waits[1], 5)
		assert.Equal(t, schedule.Combos[1].Sounds[0], "215")

		requiredSounds := []int{1, 2, 3, 212, 215}

		assert.Equal(t, requiredSounds, schedule.GetReferencedSounds())
	}
}
