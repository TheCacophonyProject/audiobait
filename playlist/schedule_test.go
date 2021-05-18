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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

const rawSchedule = `{
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
}`
const rawServerResponse = `{"messages":[],"schedule":` + rawSchedule + `,"success":true}`

var expectedSchedule = Schedule{
	Combos: []Combo{
		{
			From:    *NewTimeOfDay("19:03"),
			Every:   180,
			Until:   *NewTimeOfDay("21:30"),
			Waits:   []int{0, 5, 2},
			Sounds:  []string{"random", "same", "212"},
			Volumes: []int{2, 10, 5},
		},
		{
			From:    *NewTimeOfDay("22:00"),
			Every:   270,
			Until:   *NewTimeOfDay("02:00"),
			Waits:   []int{0},
			Sounds:  []string{"215"},
			Volumes: []int{9},
		},
	},
	PlayNights:    1,
	Description:   "Simple schedule. ",
	ControlNights: 3,
	AllSounds:     []int{1, 2, 3, 212, 215},
}

func TestBytesToSchedule(t *testing.T) {
	schedule, err := bytesToSchedule([]byte(rawSchedule))
	assert.NoError(t, err)
	assert.Equal(t, expectedSchedule, *schedule)
}

func TestServerResponseToSchedule(t *testing.T) {
	schedule, err := serverResponseToSchedule([]byte(rawServerResponse))
	assert.NoError(t, err)
	assert.Equal(t, expectedSchedule, *schedule)
}

func TestGetReferenceSounds(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3, 212, 215}, expectedSchedule.GetReferencedSounds())
	var scheduleNoRandom = Schedule{
		Combos: []Combo{
			{Sounds: []string{"same", "212"}},
			{Sounds: []string{"215"}},
		},
		AllSounds: []int{1, 2, 3, 212, 215},
	}
	expected := []int{215, 212}
	actual := scheduleNoRandom.GetReferencedSounds()
	sort.Ints(expected)
	sort.Ints(actual)
	assert.Equal(t, expected, actual)
}
