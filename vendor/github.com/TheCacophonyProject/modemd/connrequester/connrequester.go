package connrequester

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/godbus/dbus"
)

const (
	dbusPath      = "/org/cacophony/modemd"
	dbusDest      = "org.cacophony.modemd"
	methodBase    = "org.cacophony.modemd"
	wifiInterface = "wlan0"

	pingTimeout = 5
)

var (
	maxRetryInterval = time.Hour * 2
	requestinterval  = time.Second * 20
	hosts            = []string{
		"8.8.8.8",
		"8.8.4.4",
	}
)

type ConnectionRequester struct {
	stateChange  chan bool
	sendRequests bool
}

// NewConnectionRequester will return a ConnectionRequester.
// No connection will be requested until Start is called
func NewConnectionRequester() *ConnectionRequester {
	cr := &ConnectionRequester{
		stateChange:  make(chan bool, 1),
		sendRequests: false,
	}
	go cr.requestConnections()
	return cr
}

// WaitUntilUp will wait until a connection has been made returning an error
// if no connection is made in the given duration
func (cr *ConnectionRequester) WaitUntilUp(timeout time.Duration) error {
	connectionTimeout := time.After(timeout)
	for {
		if CheckConnection() {
			return nil
		}
		select {
		case <-connectionTimeout:
			return errors.New("connection failed")
		case <-time.After(time.Second):
		}
	}
}

// WaitUntilUpLoop will wait until a connection has been made returning an error
// if no connection is made.
// timeout is the time given to make a connection each try.
// retryAfter is the duration between attempts, it will multipy by 1.5 after
// each try with a maximum of 2 hours.
// maxRetries is how many times it will try to make a connection. If -1 it will
// try until a connection is made.
func (cr *ConnectionRequester) WaitUntilUpLoop(
	timeout time.Duration,
	retryAfter time.Duration,
	maxRetries int) error {
	retry := 0
	for {
		if err := cr.WaitUntilUp(timeout); err == nil {
			return nil
		}
		if maxRetries != -1 && retry >= maxRetries {
			return errors.New("no connection made")
		}
		retry++
		cr.Stop() // Stopping requesting to save power
		log.Println("connection failed. Retry in", retryAfter)
		time.Sleep(retryAfter)
		retryAfter = retryAfter * 3 / 2
		if retryAfter > maxRetryInterval {
			retryAfter = maxRetryInterval
		}
		cr.Start()
	}
}

// CheckWifiConnection will return true if the wifi is connected to the internet
func CheckWifiConnection() bool {
	return isInterfaceUp(wifiInterface) && pingAllHosts(wifiInterface)
}

func isInterfaceUp(interfaceName string) bool {
	outBytes, err := exec.Command("ip", "a", "show", interfaceName).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(outBytes), "state UP")
}

func CheckConnection() bool {
	return pingAllHosts("")
}

func pingAllHosts(interfaceName string) bool {
	pingChan := make(chan bool)
	for _, host := range hosts {
		go func(host string) {
			pingChan <- ping(interfaceName, host)
		}(host)
	}
	fails := 0
	success := false
	timeout := time.After(time.Second * pingTimeout)
	for {
		select {
		case success = <-pingChan:
			if success {
				return true
			}
			fails++
			if fails >= len(hosts) {
				return false
			}
		case <-timeout:
			return false
		}
	}
}

func ping(interfaceName string, host string) bool {
	params := []string{}
	if interfaceName != "" {
		params = []string{
			"-I",
			interfaceName,
		}
	}
	params = append(params, "-n",
		"-q",
		"-c1",
		fmt.Sprintf("-w%d", pingTimeout),
		host)
	return exec.Command("ping", params...).Run() == nil
}

// Start will start requesting for a connection to be made.
func (cr *ConnectionRequester) Start() {
	cr.stateChange <- true
	return
}

// Stop will stop requesting for a connection to be made.
func (cr *ConnectionRequester) Stop() {
	cr.stateChange <- false
	return
}

func (cr *ConnectionRequester) requestConnections() {
	for {
		newRequestTime := make(<-chan time.Time)
		if cr.sendRequests {
			if !CheckWifiConnection() {
				if err := sendOnRequest(); err != nil {
					log.Println("error with sending dbus signal: ", err)
				}
			}
			newRequestTime = time.After(requestinterval)
		}
		select {
		case cr.sendRequests = <-cr.stateChange:
		case <-newRequestTime:
		}
	}
}

func getDbusObj() (dbus.BusObject, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	obj := conn.Object(dbusDest, dbusPath)
	return obj, nil
}

func sendOnRequest() error {
	obj, err := getDbusObj()
	if err != nil {
		return err
	}
	return obj.Call(methodBase+".StayOn", 0).Store()
}
