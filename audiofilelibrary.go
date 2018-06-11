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

		if line[0] != '%' {
			parts := strings.Split(line, ":")

			if len(parts) > 1 {
				library.FilesById[parts[0]] = strings.Trim(parts[1], " ")
			}
		}
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
		_, _ = f.WriteString("#  This is a the list of all the audio files downloaded indexed by id of file")
	}

	text := "\r\n" + fileId + ": " + filename
	_, err = f.WriteString(text)

	return err
}
