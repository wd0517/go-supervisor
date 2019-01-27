package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Supervisord *SupervisorConfig `yaml:"supervisord"`
	Processes   []*ProcessConfig  `yaml:"processes`
}

type SupervisorConfig struct {
	LogPath    string `yaml:"logpath"`
	Logfile    string `yaml:"logfile"`
	Pidfile    string `yaml:"pidfile"`
	Httpserver string `yaml:"httpserver"`
	Nodaemon   bool   `yaml:"nodaemon"`
}

type ProcessConfig struct {
	Name            string  `yaml:"name"`
	Command         string  `yaml:"command"`
	Directory       string  `yaml:"directory"`
	AutoStart       bool    `yaml:"autostart"`
	AutoreReStart   bool    `yaml:"autorestart"`
	StartSecs       float64 `yaml:"startsecs"`
	StartRetries    int     `yaml:"startretries"`
	LogBackups      int     `yaml:"logbackups"`
	LogMaxMegaBytes int     `yaml:"logmaxmegabytes"`
}

func loadConfig(confFile string) (*Config, error) {
	c := &Config{Supervisord: &SupervisorConfig{}}
	c.Supervisord.Logfile = "supervisor.log"
	c.Supervisord.Pidfile = ".supervisor.pid"
	c.Supervisord.LogPath = "logs"
	c.Supervisord.Httpserver = "127.0.0.1:8080"
	c.Supervisord.Nodaemon = false

	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		return c, errors.New("Error: Config file not exists!")
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return c, errors.New("Error: Cannot load Config file correctly!")
	}
	return c, nil
}

func echoExampleConfig() {
	fmt.Fprintf(os.Stdout, `supervisord:
  logfile: "supervisor.log"
  pidfile: ".supervisor.pid"
  logpath: "logs"
  httpserver: "127.0.0.1:8001"
  nodaemon: false
processes:
- name: fileServer
  command: python -m SimpleHTTPServer 8002
  autostart: true
  autorestart: true
  startsecs: 2
  startretries: 3
  logmaxmegabytes: 1
  logbackups: 10`)
}
