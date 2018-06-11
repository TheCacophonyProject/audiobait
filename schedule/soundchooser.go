package schedule

import (
	"math/rand"
	"strconv"
	"time"
)

type SoundChooser struct {
	allSounds map[int]string //Map of sound file id (from api database) to the filename on disk
	allKeys   []int
	random    *rand.Rand
	previous  int
}

func NewSoundChooser(allSoundsMap map[int]string) *SoundChooser {
	return NewSoundChooserWithRandom(allSoundsMap, time.Now().UnixNano())
}

func NewSoundChooserWithRandom(allSoundsMap map[int]string, seed int64) *SoundChooser {
	soundChooser := SoundChooser{random: rand.New(rand.NewSource(seed)), previous: 0}
	return (&soundChooser).setAllSounds(allSoundsMap)
}

func (chooser *SoundChooser) setAllSounds(allSoundsMap map[int]string) *SoundChooser {
	chooser.allSounds = allSoundsMap

	i := 0
	chooser.allKeys = make([]int, len(chooser.allSounds))
	for key := range chooser.allSounds {
		chooser.allKeys[i] = key
		i++
	}
	return chooser
}

func (chooser *SoundChooser) returnSound(soundId int) (int, string) {
	chooser.previous = soundId
	return soundId, chooser.allSounds[soundId]
}

func (chooser *SoundChooser) ChooseSound(choice string) (int, string) {
	if choice == "random" {
		index := chooser.random.Intn(len(chooser.allKeys))
		return chooser.returnSound(chooser.allKeys[index])
	} else if choice == "same" {
		if chooser.previous != 0 {
			return chooser.returnSound(chooser.previous)
		}
	} else {
		fileId, err := strconv.Atoi(choice)
		if err == nil {
			filename := chooser.allSounds[fileId]
			if len(filename) > 0 {
				return chooser.returnSound(fileId)
			}
		}
	}
	return 0, ""
}
