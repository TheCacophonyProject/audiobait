/*
audiobait - play sounds to lure animals for The Cacophony Project API.
Copyright (C) 2021, The Cacophony Project

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
	"sync"

	"github.com/TheCacophonyProject/audiobait/audiofilelibrary"
	"github.com/TheCacophonyProject/audiobait/playlist"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const (
	dbusName = "org.cacophony.Audiobait"
	dbusPath = "/org/cacophony/Audiobait"
)

var mu = sync.RWMutex{}

type service struct {
	soundCard SoundCardPlayer
	player    *playlist.SchedulePlayer
	soundDir  string
}

func startService(soundDir string, soundCard SoundCardPlayer) (*service, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	reply, err := conn.RequestName(dbusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return nil, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, errors.New("name already taken")
	}

	s := &service{
		soundDir:  soundDir,
		soundCard: soundCard,
	}
	if err := conn.Export(s, dbusPath, dbusName); err != nil {
		return nil, err
	}
	if err := conn.Export(genIntrospectable(s), dbusPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *service) setPlayer(player *playlist.SchedulePlayer) {
	mu.Lock()
	defer mu.Unlock()
	s.player = player
}

func (s service) Play(audioFileId, volume, priority int, makeEvent bool) (bool, *dbus.Error) {
	mu.Lock()
	defer mu.Unlock()
	library, err := audiofilelibrary.OpenLibrary(s.soundDir, scheduleFilename)
	if err != nil {
		return false, dbusErr("Play", err)
	}
	path := s.soundDir + "/" + library.FilesByID[audioFileId]
	if err := s.soundCard.Play(path, volume); err != nil {
		return false, dbusErr("Play", err)
	}
	return true, nil
}

func dbusErr(name string, err error) *dbus.Error {
	if err == nil {
		return nil
	}
	return &dbus.Error{
		Name: dbusName + "." + name,
		Body: []interface{}{err.Error()},
	}
}

func genIntrospectable(v interface{}) introspect.Introspectable {
	node := &introspect.Node{
		Interfaces: []introspect.Interface{{
			Name:    dbusName,
			Methods: introspect.Methods(v),
		}},
	}
	return introspect.NewIntrospectable(node)
}

/*	//TODO
func (s *service) PlayTrigger(trigger string, volume, priority int, makeEvent bool) (bool, *dbus.Error) {
	mu.Lock()
	defer mu.Unlock()
	return true, nil
}

func (s service) Status() *dbus.Error {
	mu.Lock()
	defer mu.Unlock()
	return nil
}

func (s service) Mute(priority int) *dbus.Error {
	return nil
}

*/
