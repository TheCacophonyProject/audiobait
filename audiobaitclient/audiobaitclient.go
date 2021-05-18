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

// PlayFromId lets you make a request to audiobait to play an audio file.
// audioFileId: ID of the audio file. Audio files available and there IDs can be found using audiofilelibrary.
// volume: Volume to play the sound at from 1 to 10. Values over 10 can be used but the quality might decrease.
// priority: //TODO
// event: Event that will get logged when played. The audioFileID, volume, priority, and time will automatically get added to the event.
//        If left null no event will be logged.
func PlayFromId(audioFileId, volume, priority int, event *eventclient.Event) (played bool, err error) {
	var eventRaw []byte
	if event != nil {
		eventRaw, err = json.Marshal(event)
		if err != nil {
			return false, err
		}
	}
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

func PlayTestSound(volume int) error {
	_, err := dbusCall("PlayTestSound", volume)
	return err
}
