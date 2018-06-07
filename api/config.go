/*
event-reporter - report events to the Cacophony Project API.
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
	"errors"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServerURL  string `yaml:"server-url"`
	Group      string `yaml:"group"`
	DeviceName string `yaml:"device-name"`
	Directory  string `yaml:"directory"`
}

func (conf *Config) Validate() error {
	if conf.ServerURL == "" {
		return errors.New("server-url missing")
	}
	if conf.Group == "" {
		return errors.New("group missing")
	}
	if conf.DeviceName == "" {
		return errors.New("device-name missing")
	}
	if conf.Directory == "" {
		return errors.New("directory missing")
	}
	return nil
}

type PrivateConfig struct {
	Password string `yaml:"password"`
}

func ParseConfigFile(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(buf)
}

func ParseConfig(buf []byte) (*Config, error) {
	conf := &Config{
		Directory: "/var/spool/cptv",
	}
	if err := yaml.Unmarshal(buf, conf); err != nil {
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	return conf, nil
}

func ReadPassword(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	var conf PrivateConfig
	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return "", err
	}
	return conf.Password, nil
}

func WritePassword(filename, password string) error {
	conf := PrivateConfig{Password: password}
	buf, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0600)
}
