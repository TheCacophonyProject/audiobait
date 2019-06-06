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
	"log"
	"time"

	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/host"
)

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

type Args struct {
	ConfigFile   string `arg:"-c,--config" help:"path to configuration file"`
	Timestamps   bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
	RestartModem bool   `arg:"-r,--restart" help:"cycle the power to the USB port"`
}

func (Args) Version() string {
	return version
}

func procArgs() Args {
	args := Args{
		ConfigFile: "/etc/cacophony/modemd.yaml",
	}
	arg.MustParse(&args)
	return args
}

var version = "<not set>"

func runMain() error {
	args := procArgs()
	if !args.Timestamps {
		log.SetFlags(0) // Removes default timestamp flag
	}
	log.Printf("running version: %s", version)

	log.Print("init gpio")
	if _, err := host.Init(); err != nil {
		return err
	}

	conf, err := ParseModemdConfig(args.ConfigFile)
	if err != nil {
		return err
	}

	log.Printf("%+v\n", conf)

	mc := ModemController{
		StartTime:         time.Now(),
		ModemsConfig:      conf.ModemsConfig,
		TestHosts:         conf.TestHosts,
		TestInterval:      conf.TestInterval,
		PowerPin:          conf.PowerPin,
		InitialOnTime:     conf.InitialOnTime,
		FindModemTime:     conf.FindModemTime,
		ConnectionTimeout: conf.ConnectionTimeout,
		PingWaitTime:      conf.PingWaitTime,
		PingRetries:       conf.PingRetries,
		RequestOnTime:     conf.RequestOnTime,
	}

	log.Println("starting dbus service")
	if err := startService(&mc); err != nil {
		return err
	}

	if mc.ShouldBeOff() || args.RestartModem {
		log.Println("powering off USB modem")
		mc.SetModemPower(false)
	}

	for {
		if mc.ShouldBeOff() {
			log.Println("waiting until modem should be powered on")
			for mc.ShouldBeOff() {
				time.Sleep(time.Second)
			}
		}

		log.Println("powering on USB modem")
		mc.SetModemPower(true)

		log.Println("finding USB modem")
		for !mc.ShouldBeOff() {
			if mc.FindModem() {
				log.Printf("found modem %s\n", mc.Modem.Name)
				break
			}
			log.Println("no USB modem found")
			mc.CycleModemPower()
		}

		if mc.Modem != nil {
			log.Println("waiting for modem to connect to a network")
			connected, err := mc.WaitForConnection()
			if err != nil {
				return err
			}
			if connected {
				log.Println("modem has connected to a network")
				for {
					if mc.PingTest() {
						log.Println("ping test passed")
					} else {
						log.Println("ping test failed")
						break
					}
					if !mc.WaitForNextPingTest() {
						break
					}
				}
			} else {
				log.Println("modem failed to connect to a network")
			}
		}

		mc.Modem = nil

		if mc.ShouldBeOff() {
			log.Println("modem should no longer be on")
		}

		log.Println("powering off USB modem")
		mc.SetModemPower(false)
	}

	return nil
}
