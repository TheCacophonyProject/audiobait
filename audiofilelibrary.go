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
	"errors"
	"io/ioutil"
	"log"
	"strings"
)

type AudioFileLibrary struct {
	soundsDirectory string
	FilesById       map[string]string
}

func OpenLibrary(soundsDirectory string) *AudioFileLibrary {
	return (&AudioFileLibrary{}).openLibrary(soundsDirectory)
}

// Take a file name and extract an ID from it.
// The file names are similar to "bellbird-6.mp3".  But they could be like "SI Kaka-SI Kaka-17.mp3"
// Need to handle cases like "schedule.json" i.e. the file is not an audio file.
func extractIDFromFileName(fileName string) (string, error) {

	lastIndex := strings.LastIndex(fileName, "-")
	if lastIndex == -1 || lastIndex+1 >= len(fileName) {
		return "", errors.New("Skipping file with name " + fileName)
	}

	fileIDWithExtension := fileName[lastIndex+1:]
	if len(fileIDWithExtension) == 0 {
		return "", errors.New("Skipping file with name " + fileName)
	}

	parts := strings.Split(fileIDWithExtension, ".")
	if len(parts) != 2 {
		return "", errors.New("Skipping file with name " + fileName)
	}

	return parts[0], nil

}

// Read the audio directory.  Construct a map of file IDs to file names.
func (library *AudioFileLibrary) openLibrary(soundsDirectory string) *AudioFileLibrary {

	log.Println("Reading audio directory")
	library.soundsDirectory = soundsDirectory
	library.FilesById = make(map[string]string)

	files, err := ioutil.ReadDir(soundsDirectory)
	if err != nil {
		log.Println("Error reading audio directory", err)
		return library
	}

	// Get IDs from the filename.
	for _, file := range files {
		fileID, err := extractIDFromFileName(file.Name())
		if err != nil {
			log.Println(err.Error())
			continue
		} else {
			library.FilesById[fileID] = file.Name()
		}

	}

	log.Println("Audio directory read successfully")
	return library
}

func (library *AudioFileLibrary) GetFileNameOnDisk(fileId string) (string, bool) {
	filename, exists := library.FilesById[fileId]
	return filename, exists
}
