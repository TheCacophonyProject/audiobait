package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os/exec"
	"strings"
	"time"
)

type Modem struct {
	Name          string
	Netdev        string
	VendorProduct string
}

// NewModem return a new modem from the config
func NewModem(config ModemConfig) *Modem {
	m := &Modem{
		Name:          config.Name,
		Netdev:        config.Netdev,
		VendorProduct: config.VendorProduct,
	}
	return m
}

// PingTest will try connecting to one of the provides hosts
func (m *Modem) PingTest(timeoutSec int, retries int, hosts []string) bool {
	for i := retries; i > 0; i-- {
		for _, host := range hosts {
			cmd := exec.Command(
				"ping",
				"-I",
				m.Netdev,
				"-n",
				"-q",
				"-c1",
				fmt.Sprintf("-w%d", timeoutSec),
				host)
			if err := cmd.Run(); err == nil {
				return true
			}
		}
		if i > 1 {
			log.Printf("ping test failed. %d more retries\n", i-1)
		}
	}
	return false
}

// IsDefaultRoute will check if the USB modem is connected
func (m *Modem) IsDefaultRoute() (bool, error) {
	outByte, err := exec.Command("ip", "route").Output()
	if err != nil {
		return false, err
	}
	out := string(outByte)
	lines := strings.Split(out, "\n")
	search := fmt.Sprintf(" dev %s ", m.Netdev)
	for _, line := range lines {
		if strings.HasPrefix(line, "default") && strings.Contains(line, search) {
			return true, nil
		}
	}
	return false, nil
}

type responseHolder struct {
	response
}
type response struct {
	SignalIcon int
}

func (m *Modem) WaitForConnection(timeout int) (bool, error) {
	start := time.Now()
	for {
		def, err := m.IsDefaultRoute()
		if err != nil {
			return false, err
		}
		if def {
			return true, nil
		}
		if time.Now().Sub(start) > time.Second*time.Duration(timeout) {
			return false, nil
		}
		time.Sleep(time.Second)
	}
}

func (m *Modem) signalStrengthHuawei() (int, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return 0, err
	}
	client := &http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * 10,
	}
	_, err = client.Get("http://192.168.8.1") // Goto homepage first to get cookie for API request
	if err != nil {
		return 0, err
	}
	resp, err := client.Get("http://192.168.8.1" + "/api/monitoring/status")
	if err != nil {
		return 0, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var r responseHolder
	err = xml.Unmarshal(bodyBytes, &r)
	if err != nil {
		return 0, err
	}
	return r.SignalIcon, nil
}
