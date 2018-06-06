package schedule

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var soundChooserFiles = map[int]string {
	1 : "squeal",
	3 : "beep",
	4 : "tweet",
}

func TestSoundChooserRandom(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 1)
	soundId, soundName := chooser.ChooseSound("random")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 4)
	assert.Equal(t, soundName, "tweet")

	soundId, soundName = chooser.ChooseSound("random")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 1)
	assert.Equal(t, soundName, "squeal")
}

func TestSoundChooserSame(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 2)

	soundId, soundName := chooser.ChooseSound("random")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("same")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("random")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 1)
	assert.Equal(t, soundName, "squeal")
}

func TestSoundChooseByFileId(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 3)

	soundId, soundName := chooser.ChooseSound("3")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("same")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("2")
	fmt.Printf("Selected sound to play %d: %s\n", soundId, soundName)
	assert.Equal(t, soundId, 0)
	assert.Equal(t, soundName, "")
}