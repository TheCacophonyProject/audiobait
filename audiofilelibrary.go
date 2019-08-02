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
	"path/filepath"
	"strconv"
	"strings"
)

// AudioFileLibrary is a structure to hold info on the files in the audio directory.
type AudioFileLibrary struct {
	soundsDirectory string
	FilesByID       map[int]string
}

// Take a file name and extract an ID from it.
// The file names are similar to "bellbird-6.mp3".  But they could be like "SI Kaka-SI Kaka-17.mp3"
// Need to handle cases like "schedule.json" i.e. the file is not an audio file.
func extractIDFromFileName(fileName string) (int, error) {

	lastIndex := strings.LastIndex(fileName, "-")
	if lastIndex < 0 {
		return -1, errors.New("Skipping file with name " + fileName)
	}

	fileIDWithExtension := fileName[lastIndex+1:]
	fileIDStr := strings.TrimSuffix(fileIDWithExtension, filepath.Ext(fileIDWithExtension))
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return -1, err
	}

	return fileID, nil
}

// OpenLibrary reads the audio directory.  And constructs a map of file IDs to file names.
func OpenLibrary(soundsDirectory string) *AudioFileLibrary {

	log.Println("Reading audio directory")
	library := &AudioFileLibrary{
		soundsDirectory : soundsDirectory,
		FilesByID : make(map[int]string),
	}

	files, err := ioutil.ReadDir(soundsDirectory)
	if err != nil {
		log.Println("Error reading audio directory", err)
		return library
	}

	// Get IDs from the filenames.
	for _, file := range files {
		fileID, err := extractIDFromFileName(file.Name())
		if err == nil {
			library.FilesByID[fileID] = file.Name()
		}
	}

	log.Println("Audio directory read successfully")
	return library
}

// GetFileNameOnDisk takes a fileID and returns it's name, which was previously read from disk.
func (library *AudioFileLibrary) GetFileNameOnDisk(fileID int) (string, bool) {
	filename, exists := library.FilesByID[fileID]
	return filename, exists
}

// MakeFileName takes some file details retrieved from the API server and constructs
// the name we want the file to have when written to disk.
func MakeFileName(APIOriginalFileName string, APIFileName string, fileID int) string {

	// fileNameParts := strings.Split(apiOriginalFileName, ".")
	// fileExt := ""
	// if len(fileNameParts) > 1 {
	// 	fileExt = "." + fileNameParts[len(fileNameParts)-1]
	// }
	fileExt := filepath.Ext(APIOriginalFileName)
	return APIFileName + "-" + strconv.Itoa(fileID) + fileExt

}
