package main

import (
	"errors"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v1"
)

type Config struct {
	AudioDir      string
	Card          int
	VolumeControl string
	WindowStart   time.Time
	WindowEnd     time.Time
	Play          PlayConfig
}

type PlayConfig struct {
	File        string        `yaml:"string"`
	BurstRepeat int           `yaml:"burst-repeat"`
	IntraSleep  time.Duration `yaml:"intra-sleep"`
	InterSleep  time.Duration `yaml:"inter-sleep"`
}

type rawConfig struct {
	AudioDir      string        `yaml:"audio-directory"`
	Card          int           `yaml:"card"`
	VolumeControl string        `yaml:"volume-control"`
	WindowStart   string        `yaml:"window-start"`
	WindowEnd     string        `yaml:"window-end"`
	Play          rawPlayConfig `yaml:"play"`
}

type rawPlayConfig struct {
	File        string `yaml:"file"`
	BurstRepeat int    `yaml:"burst-repeat"`
	IntraSleep  string `yaml:"intra-sleep"`
	InterSleep  string `yaml:"inter-sleep"`
}

const fmtTimeOnly = "15:04"

func ParseConfigFile(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(buf)
}

func ParseConfig(buf []byte) (*Config, error) {
	var raw rawConfig
	err := yaml.Unmarshal(buf, &raw)
	if err != nil {
		return nil, err
	}

	conf := &Config{
		AudioDir:      raw.AudioDir,
		Card:          raw.Card,
		VolumeControl: raw.VolumeControl,
		Play: PlayConfig{
			File:        raw.Play.File,
			BurstRepeat: raw.Play.BurstRepeat,
		},
	}

	conf.WindowStart, err = time.Parse(fmtTimeOnly, raw.WindowStart)
	if err != nil {
		return nil, errors.New("invalid window-start")
	}

	conf.WindowEnd, err = time.Parse(fmtTimeOnly, raw.WindowEnd)
	if err != nil {
		return nil, errors.New("invalid window-end")
	}

	conf.Play.IntraSleep, err = time.ParseDuration(raw.Play.IntraSleep)
	if err != nil {
		return nil, errors.New("invalid intra-sleep")
	}

	conf.Play.InterSleep, err = time.ParseDuration(raw.Play.InterSleep)
	if err != nil {
		return nil, errors.New("invalid inter-sleep")
	}

	return conf, nil
}
