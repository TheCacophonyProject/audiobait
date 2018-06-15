// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package playlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var soundChooserFiles = map[int]string{
	1: "squeal",
	3: "beep",
	4: "tweet",
}

func TestSoundChooserRandom(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 1)
	soundId, soundName := chooser.ChooseSound("random")
	assert.Equal(t, soundId, 4)
	assert.Equal(t, soundName, "tweet")

	soundId, soundName = chooser.ChooseSound("random")
	assert.Equal(t, soundId, 1)
	assert.Equal(t, soundName, "squeal")
}

func TestSoundChooserSame(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 2)

	soundId, soundName := chooser.ChooseSound("random")
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("same")
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("random")
	assert.Equal(t, soundId, 1)
	assert.Equal(t, soundName, "squeal")
}

func TestSoundChooserCantUnderstand(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 2)

	soundId, _ := chooser.ChooseSound("unparsable")
	assert.Equal(t, soundId, 0)

	// there is no current sound to repeat
	soundId, _ = chooser.ChooseSound("same")
	assert.Equal(t, soundId, 0)

	//sound doesn't exist
	indexDoesntExist := "512"
	soundId, _ = chooser.ChooseSound(indexDoesntExist)
	assert.Equal(t, soundId, 0)
}

func TestSoundChooseByFileId(t *testing.T) {
	chooser := NewSoundChooserWithRandom(soundChooserFiles, 3)

	soundId, soundName := chooser.ChooseSound("3")
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("same")
	assert.Equal(t, soundId, 3)
	assert.Equal(t, soundName, "beep")

	soundId, soundName = chooser.ChooseSound("2")
	assert.Equal(t, soundId, 0)
	assert.Equal(t, soundName, "")
}
