// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/godbus/dbus"
)

// AudioBaitEventRecorder uses the event api to record that audioBait was played at a particular time.
type AudioBaitEventRecorder struct {
}

func (er AudioBaitEventRecorder) OnAudioBaitPlayed(ts time.Time, fileId int, volume int) {
	eventDetails := map[string]interface{}{
		"description": map[string]interface{}{
			"type": "audioBait",
			"details": map[string]interface{}{
				"fileId": fileId,
				"volume": volume,
			},
		},
	}
	detailsJSON, err := json.Marshal(&eventDetails)
	if err != nil {
		log.Printf("Could not log audiobait played: %s", err)
		return
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Printf("Could not log audiobait played: %s", err)
		return
	}

	obj := conn.Object("org.cacophony.Events", "/org/cacophony/Events")
	call := obj.Call("org.cacophony.Events.Queue", 0, detailsJSON, ts.UnixNano())
	if call.Err != nil {
		log.Printf("Could not log audiobait played: %s", call.Err)
		return
	}
}
