/*
audiobaitclient - client for playing audiobait sounds
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
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package audiobaitclient

import (
	"errors"

	"github.com/godbus/dbus"
)

func Play(audioFileId, volume, priority int, makeEvent bool) (played bool, err error) {
	data, err := dbusCall("Play", audioFileId, volume, priority, makeEvent)
	if err != nil {
		return false, err
	}
	if len(data) != 1 {
		return false, errors.New("error playing sound")
	}
	played, ok := data[0].(bool)
	if !ok {
		return false, errors.New("error playing sound")
	}
	return played, nil

}

func dbusCall(method string, params ...interface{}) ([]interface{}, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	obj := conn.Object("org.cacophony.Audiobait", "/org/cacophony/Audiobait")
	call := obj.Call(method, 0, params...)
	return call.Body, call.Err
}
