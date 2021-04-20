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
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/TheCacophonyProject/go-api"
)

const (
	ScheduleFilename = "schedule.json"
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
	Trigger string
}

func SaveScheduleIfNew(audioDir string, newSchedule *Schedule) (new bool, err error) {
	oldSchedule, err := LoadScheduleFromDisk(audioDir)
	if err != nil {
		log.Printf("error loading old schedule so saving new schedule '%v'\n", err)
		return true, saveScheduleToDisk(audioDir, newSchedule)
	}
	if !reflect.DeepEqual(oldSchedule, newSchedule) {
		log.Println("saving new schedule to disk")
		return true, saveScheduleToDisk(audioDir, newSchedule)
	}
	return false, nil
}

func saveScheduleToDisk(audioDir string, schedule *Schedule) error {
	marshedSchedule, err := json.Marshal(schedule)
	if err != nil {
		return err
	}
	filename := filepath.Join(audioDir, ScheduleFilename)
	return ioutil.WriteFile(filename, marshedSchedule, 0644)
}

func GetScheduleFromAPI(api *api.CacophonyAPI) (*Schedule, error) {
	responseBytes, err := api.GetSchedule()
	if err != nil {
		return nil, err
	}
	return serverResponseToSchedule(responseBytes)
}

func serverResponseToSchedule(bytes []byte) (*Schedule, error) {
	type scheduleServerResponse struct {
		Schedule Schedule
	}
	var sr scheduleServerResponse
	if err := json.Unmarshal(bytes, &sr); err != nil {
		return nil, err
	}
	return &sr.Schedule, nil
}

func LoadScheduleFromDisk(audioDir string) (*Schedule, error) {
	rawData, err := ioutil.ReadFile(path.Join(audioDir, ScheduleFilename))
	if err != nil {
		return nil, err
	}
	return bytesToSchedule(rawData)
}

func bytesToSchedule(bytes []byte) (*Schedule, error) {
	var s Schedule
	if err := json.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func ParseServerResponse(jsonAsString string, schedule *Schedule) error {
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
