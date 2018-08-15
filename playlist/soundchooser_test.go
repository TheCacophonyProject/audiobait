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
	"testing"

	"github.com/stretchr/testify/assert"
)

var soundChooserFiles = map[int]string{
	1: "squeal",
	3: "beep",
	4: "tweet",
}

// These tests have been disabled because they sometimes fail,
// due to even seeded random choosing different values.  They pass atleast 75%
// of the time so may be userful for development.

// func TestSoundChooserRandom(t *testing.T) {
// 	chooser := NewSoundChooserWithRandom(soundChooserFiles, 1)
// 	soundId, soundName := chooser.ChooseSound("random")
// 	assert.Equal(t, soundId, 4)
// 	assert.Equal(t, soundName, "tweet")

// 	soundId, soundName = chooser.ChooseSound("random")
// 	assert.Equal(t, soundId, 1)
// 	assert.Equal(t, soundName, "squeal")
// }

// func TestSoundChooserSame(t *testing.T) {
// 	chooser := NewSoundChooserWithRandom(soundChooserFiles, 2)

// 	soundId, soundName := chooser.ChooseSound("random")
// 	assert.Equal(t, soundId, 3)
// 	assert.Equal(t, soundName, "beep")

// 	soundId, soundName = chooser.ChooseSound("same")
// 	assert.Equal(t, soundId, 3)
// 	assert.Equal(t, soundName, "beep")

// 	soundId, soundName = chooser.ChooseSound("random")
// 	assert.Equal(t, soundId, 1)
// 	assert.Equal(t, soundName, "squeal")
// }

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
