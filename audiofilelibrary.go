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

package main

import (
	"strings"
	"io/ioutil"
	"log"
)

type AudioFileLibrary struct {
	soundsDirectory string
	FilesById       map[string]string
}

func OpenLibrary(soundsDirectory string) *AudioFileLibrary {
	return (&AudioFileLibrary{}).openLibrary(soundsDirectory)
}

func (library *AudioFileLibrary) openLibrary(soundsDirectory string) *AudioFileLibrary {
	library.soundsDirectory = soundsDirectory
	library.FilesById = make(map[string]string)

	files, err := ioutil.ReadDir(soundsDirectory)
	if err != nil {
		log.Println("Error reading audio directory", err)
		return library
	}
	// Get IDs from the filename.
	// TODO: Make this more robust.  Could be more than one hyphen in file name.  And names like schedule.json
	for _, file := range files {
		log.Println(file.Name())
		parts := strings.Split(file.Name(), "-")
		if len(parts) != 2 {
			log.Println("Could not parse file with name", file.Name())
			continue
		}
		log.Println(parts)
		fileID := strings.Split(parts[1], ".")[0]
		log.Println(fileID)
		library.FilesById[fileID] = file.Name()
	}

	return library
}

func (library *AudioFileLibrary) GetFileNameOnDisk(fileId string) (string, bool) {
	filename, exists := library.FilesById[fileId]
	return filename, exists
}
