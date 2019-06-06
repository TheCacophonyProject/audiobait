/*
modemd - Communicates with USB modems
Copyright (C) 2019, The Cacophony Project

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

package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

type ModemController struct {
	StartTime         time.Time
	Modem             *Modem
	ModemsConfig      []ModemConfig
	TestHosts         []string
	TestInterval      time.Duration
	PowerPin          string
	InitialOnTime     time.Duration
	FindModemTime     time.Duration // Time in seconds after USB powered on for the modem to be found
	ConnectionTimeout time.Duration // Time in seconds for modem to make a connection to the network
	PingWaitTime      time.Duration
	PingRetries       int
	RequestOnTime     time.Duration // Time the modem will stay on in seconds after a request was made

	lastOnRequestTime time.Time
}

func (mc *ModemController) NewOnRequest() {
	log.Println("stay on request")
	mc.lastOnRequestTime = time.Now()
}

func (mc *ModemController) FindModem() bool {
	timeout := time.After(mc.FindModemTime)
	for {
		select {
		case <-timeout:
			return false
		case <-time.After(time.Second):
			for _, modemConfig := range mc.ModemsConfig {
				cmd := exec.Command("lsusb", "-d", modemConfig.VendorProduct)
				if err := cmd.Run(); err == nil {
					mc.Modem = NewModem(modemConfig)
					return true
				}
			}
		}
	}
}

func (mc *ModemController) SetModemPower(on bool) error {
	pin := gpioreg.ByName(mc.PowerPin)
	if on {
		if err := pin.Out(gpio.High); err != nil {
			return fmt.Errorf("failed to set modem power pin high: %v", err)
		}
	} else {
		if err := pin.Out(gpio.Low); err != nil {
			return fmt.Errorf("failed to set modem power pin low: %v", err)
		}
		time.Sleep(time.Second * 5)
	}
	return nil
}

func (mc *ModemController) CycleModemPower() error {
	if err := mc.SetModemPower(false); err != nil {
		return err
	}
	return mc.SetModemPower(true)
}

// WaitForConnection will return false if no connection is made before either
// it timeouts or the modem should no longer be powered.
func (mc *ModemController) WaitForConnection() (bool, error) {
	timeout := time.After(mc.ConnectionTimeout)
	for {
		select {
		case <-timeout:
			return false, nil
		case <-time.After(time.Second):
			def, err := mc.Modem.IsDefaultRoute()
			if err != nil {
				return false, err
			}
			if def {
				return true, nil
			}
		}
	}
}

// ShouldBeOff will look at the following factors to determine if the modem should be off.
// - InitialOnTime: Modem should be on for a set amount of time at the start.
// - LastOnRequest: Check if the last "StayOn" request was less than 'RequestOnTime' ago.
// - OnWindow: //TODO
func (mc *ModemController) ShouldBeOff() bool {
	if !timeoutCheck(mc.StartTime, mc.InitialOnTime) {
		return false
	}

	if !timeoutCheck(mc.lastOnRequestTime, mc.RequestOnTime) {
		return false
	}

	return true
}

// WaitForNextPingTest will return false if when waiting ShouldBeOff returns
// true, otherwise will return true after waiting.
func (mc *ModemController) WaitForNextPingTest() bool {
	timeout := time.After(mc.TestInterval)
	for {
		select {
		case <-timeout:
			return true
		case <-time.After(time.Second):
			if mc.ShouldBeOff() {
				return false
			}
		}
	}
}

func (mc *ModemController) PingTest() bool {
	seconds := int(mc.PingWaitTime / time.Second)
	return mc.Modem.PingTest(seconds, mc.PingRetries, mc.TestHosts)
}

func timeoutCheck(startTime time.Time, timeout time.Duration) bool {
	return time.Now().Sub(startTime) > timeout
}
