/*
audiobait - play sounds to lure animals for the CacophonyProject API.
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

package api

import (
	"math/rand"
	"time"
)

const (
	chars       = "abcdefghijklmnopqrstuvwxyz0123456789"
	charIdxBits = 6                  // 6 bits to represent a char index
	charIdxMax  = 63 / charIdxBits   // # of char indices fitting in 63 bits
	charIdxMask = 1<<charIdxBits - 1 // All 1-bits, as many as charIdxBits
)

var randSrc = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
	b := make([]byte, n)
	// A randSrc.Int63() generates 63 random bits, enough for charIdxMax characters!
	for i, cache, remain := n-1, randSrc.Int63(), charIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSrc.Int63(), charIdxMax
		}
		if idx := int(cache & charIdxMask); idx < len(chars) {
			b[i] = chars[idx]
			i--
		}
		cache >>= charIdxBits
		remain--
	}
	return string(b)
}
