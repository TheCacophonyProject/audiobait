/*
audiobait - play sounds to lure animals for The Cacophony Project API.
Copyright (C) 2018, The Cacophony Project

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

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const httpTimeout = 60 * time.Second

// NewAPI creates a CacophonyAPI instance and obtains a fresh JSON Web
// Token. If no password is given then the device is registered.
func NewAPI(serverURL, group, deviceName, password string) (*CacophonyAPI, error) {
	api := &CacophonyAPI{
		serverURL:  serverURL,
		group:      group,
		deviceName: deviceName,
		password:   password,
	}
	err := api.newToken()
	if err != nil {
		return nil, err
	}
	return api, nil
}

type CacophonyAPI struct {
	serverURL      string
	group          string
	deviceName     string
	password       string
	token          string
	justRegistered bool
}

func (api *CacophonyAPI) Password() string {
	return api.password
}

func (api *CacophonyAPI) JustRegistered() bool {
	return api.justRegistered
}

func (api *CacophonyAPI) newToken() error {
	if api.password == "" {
		return errors.New("no password set")
	}
	payload, err := json.Marshal(map[string]string{
		"devicename": api.deviceName,
		"password":   api.password,
	})
	if err != nil {
		return err
	}
	postResp, err := http.Post(
		api.serverURL+"/authenticate_device",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	defer postResp.Body.Close()

	var resp tokenResponse
	d := json.NewDecoder(postResp.Body)
	if err := d.Decode(&resp); err != nil {
		return fmt.Errorf("decode: %v", err)
	}
	if !resp.Success {
		return fmt.Errorf("registration failed: %v", resp.message())
	}
	api.token = resp.Token
	return nil
}

type tokenResponse struct {
	Success  bool
	Messages []string
	Token    string
}

func (r *tokenResponse) message() string {
	if len(r.Messages) > 0 {
		return r.Messages[0]
	}
	return "unknown"
}

func (api *CacophonyAPI) getFileFromJWT(jwt, path string) error {
	// Create the file

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(api.serverURL + "/api/v1/signedUrl?jwt=" + jwt)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// GetFileDetails will download the file details from the files api.  This can then be parsed into
// DownloadFile to download the file
func (api *CacophonyAPI) GetFileDetails(fileID int) (*FileResponse, error) {
	buf := new(bytes.Buffer)

	req, err := http.NewRequest("GET", api.serverURL+"/api/v1/files/"+strconv.Itoa(fileID), buf)
	req.Header.Set("Authorization", api.token)
	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fr FileResponse
	d := json.NewDecoder(resp.Body)
	if err := d.Decode(&fr); err != nil {
		return &fr, err
	}
	return &fr, nil
}

// DownloadFile will take the file details from GetFileDetails and download the file to a specified path
func (api *CacophonyAPI) DownloadFile(fileResponse *FileResponse, filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return err
	}

	return api.getFileFromJWT(fileResponse.Jwt, filePath)
}

type FileResponse struct {
	File FileInfo
	Jwt  string
}

type FileInfo struct {
	Details FileDetails
	Type    string
}

type FileDetails struct {
	Name         string
	OriginalName string
}

func (api *CacophonyAPI) ReportEvent(jsonDetails []byte, times []time.Time) error {
	// Deserialise the JSON event details into a map.
	var details map[string]interface{}
	err := json.Unmarshal(jsonDetails, &details)
	if err != nil {
		return err
	}

	// Convert the event times for sending and add to the map to send.
	dateTimes := make([]string, 0, len(times))
	for _, t := range times {
		dateTimes = append(dateTimes, formatTimestamp(t))
	}
	details["dateTimes"] = dateTimes

	// Serialise the map back to JSON for sending.
	jsonAll, err := json.Marshal(details)
	if err != nil {
		return err
	}

	// Prepare request.
	req, err := http.NewRequest("POST", api.serverURL+"/api/v1/events", bytes.NewReader(jsonAll))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", api.token)

	// Send.
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return temporaryError(err)
	}
	defer resp.Body.Close()

	if !isHTTPSuccess(resp.StatusCode) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return temporaryError(fmt.Errorf("request failed (%d) and body read failed: %v", resp.StatusCode, err))
		}
		return &Error{
			message:   fmt.Sprintf("HTTP request failed (%d): %s", resp.StatusCode, body),
			permanent: isHTTPClientError(resp.StatusCode),
		}
	}
	return nil
}

// Error is returned by API calling methods. As well as an error
// message, it includes whether the error is permanent or not.
type Error struct {
	message   string
	permanent bool
}

// Error implemented the error interface.
func (e *Error) Error() string {
	return e.message
}

// Permanent returns true if the error is permanent. Operations
// resulting in non-permanent/temporary errors may be retried.
func (e *Error) Permanent() bool {
	return e.permanent
}

// IsPermanentError examines the supplied error and returns true if it
// is permanent.
func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*Error); ok {
		return apiErr.Permanent()
	}
	// non-Errors are considered permanent.
	return true
}

func isHTTPSuccess(code int) bool {
	return code >= 200 && code < 300
}

func isHTTPClientError(code int) bool {
	return code >= 400 && code < 500
}

func temporaryError(err error) *Error {
	return &Error{message: err.Error(), permanent: false}
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// GetSchedule will get the audio schedule
func (api *CacophonyAPI) GetSchedule() ([]byte, error) {
	req, err := http.NewRequest("GET", api.serverURL+"/api/v1/schedules", nil)
	req.Header.Set("Authorization", api.token)
	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
