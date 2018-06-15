// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v1"
)

type AudioConfig struct {
	AudioDir      string `yaml:"audio-directory"`
	Card          int    `yaml:"card"`
	VolumeControl string `yaml:"volume-control"`
}

func ParseConfigFile(filename string) (*AudioConfig, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var audioConfig AudioConfig
	err = yaml.Unmarshal(buf, &audioConfig)
	if err != nil {
		return nil, err
	}
	return &audioConfig, nil
}
