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
	"encoding/json"
	"errors"
	"log"

	"github.com/TheCacophonyProject/event-reporter/eventclient"
	"github.com/godbus/dbus"
)

// Can be mocked for testing
var dbusCall = func(method string, params ...interface{}) ([]interface{}, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	obj := conn.Object("org.cacophony.Audiobait", "/org/cacophony/Audiobait")
	call := obj.Call(method, 0, params...)
	return call.Body, call.Err
}

var ErrorParsingOutput = errors.New("error with parsing dbus output")

func PlayFromId(audioFileId, volume, priority int, event *eventclient.Event) (played bool, err error) {
	eventRaw := []byte{}
	if event != nil {
		eventRaw, err = json.Marshal(event)
		if err != nil {
			return false, err
		}
	}
	log.Println(eventRaw)
	data, err := dbusCall("PlayFromId", audioFileId, volume, priority, string(eventRaw))
	if err != nil {
		return false, err
	}
	if len(data) != 1 {
		return false, ErrorParsingOutput
	}
	played, ok := data[0].(bool)
	if !ok {
		return false, ErrorParsingOutput
	}
	return played, nil
}
