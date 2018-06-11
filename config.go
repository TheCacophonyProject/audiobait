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
