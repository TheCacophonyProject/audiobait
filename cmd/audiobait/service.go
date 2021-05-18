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
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"sync"

	"github.com/TheCacophonyProject/event-reporter/eventclient"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const (
	dbusName = "org.cacophony.Audiobait"
	dbusPath = "/org/cacophony/Audiobait"
)

var mu = sync.RWMutex{}

type service struct {
	player player
}

func startService(player player) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	reply, err := conn.RequestName(dbusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("name already taken")
	}
	s := &service{
		player: player,
	}
	if err := conn.Export(s, dbusPath, dbusName); err != nil {
		return err
	}
	if err := conn.Export(genIntrospectable(s), dbusPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return err
	}
	return nil
}

func (s service) PlayFromId(fileId, volume, priority int, eventRaw string) (bool, *dbus.Error) {
	mu.Lock()
	defer mu.Unlock()
	var event *eventclient.Event
	if len(eventRaw) != 0 {
		if err := json.Unmarshal([]byte(eventRaw), &event); err != nil {
			return false, dbusErr(err)
		}
	}
	played, err := s.player.PlayFromId(fileId, volume, priority, event)
	if err != nil {
		return played, dbusErr(err)
	}
	return played, nil
}

func (s service) PlayTestSound(volume int) *dbus.Error {
	mu.Lock()
	defer mu.Unlock()
	err := s.player.PlayTestSound(volume)
	if err != nil {
		return dbusErr(err)
	}
	return nil
}

func dbusErr(err error) *dbus.Error {
	if err == nil {
		return nil
	}
	return &dbus.Error{
		Name: dbusName + "." + getCallerName(),
		Body: []interface{}{err.Error()},
	}
}

func getCallerName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n == 0 {
		return ""
	}
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return ""
	}
	funcNames := strings.Split(caller.Name(), ".")
	return funcNames[len(funcNames)-1]
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
