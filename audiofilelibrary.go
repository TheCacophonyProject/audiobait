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
	"bufio"
	"log"
	"os"
	"strings"
)

type AudioFileLibrary struct {
	filePath  string
	FilesById map[string]string
}

func OpenLibrary(filePath string) *AudioFileLibrary {
	return (&AudioFileLibrary{}).openLibrary(filePath)
}

func (library *AudioFileLibrary) openLibrary(filePath string) *AudioFileLibrary {
	library.filePath = filePath
	library.FilesById = make(map[string]string)

	// Open the file and scan it.
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error loading audio library %s", err)
		return library
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 0 && line[0] != '%' {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) > 1 {
				library.FilesById[parts[0]] = strings.Trim(parts[1], " ")
			}
		}
	}
	if scanner.Err() != nil {
		log.Printf("Error loading audio library %s", scanner.Err())
	}

	return library
}

func (library *AudioFileLibrary) AddFile(fileId, filename string) error {
	firstItem := len(library.FilesById) == 0

	library.FilesById[fileId] = filename

	f, err := os.OpenFile(library.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if firstItem {
		_, _ = f.WriteString("\n#  This is a the list of all the audio files downloaded indexed by id of file")
	}

	text := "\n" + fileId + ": " + filename
	_, err = f.WriteString(text)

	return err
}

func (library *AudioFileLibrary) GetFileNameOnDisk(fileId string) (string, bool) {
	filename, exists := library.FilesById[fileId]
	return filename, exists
}
