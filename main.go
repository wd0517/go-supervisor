package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagHelp            bool
	flagNodaemon        bool
	flagEchoExampleConf bool
	flagSignal          string
	flagConfFile        string
)

var (
	supervisor *Supervisor
)

func init() {
	flag.BoolVar(&flagHelp, "h", false, "this help")
	flag.BoolVar(&flagNodaemon, "nodaemon", false, "run in the foreground")
	flag.BoolVar(&flagEchoExampleConf, "echo_conf", false, "echo example config")
	flag.StringVar(&flagSignal, "s", "", "send `signal` to process: shutdown, stop")
	flag.StringVar(&flagConfFile, "c", "config.yml", "specify config file")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if flagHelp {
		flag.Usage()
		return
	}

	if flagEchoExampleConf {
		echoExampleConfig()
		return
	}

	supervisorConfig, err := loadConfig(flagConfFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	initSupervisor(supervisorConfig)

	if flagNodaemon {
		runSupervisor()
	} else {
		runAsDaemon(runSupervisor)
	}
}

func initSupervisor(config *Config) {
	supervisor = &Supervisor{config: config.Supervisord, Pros: make(map[string]*SubProcess)}

	for _, proConfig := range config.Processes {
		supervisor.Pros[proConfig.Name] = &SubProcess{
			Name:   proConfig.Name,
			State:  STOPPED,
			config: proConfig,
		}
	}
}

func runSupervisor() {
	supervisor.runforever()
	if supervisor.config.Httpserver != "" {
		startWebServer(supervisor.config.Httpserver)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Go-Supervisor version: 1.0.0
Usage: [-h] [-s signal] [-nodaemon] [-c config] [-echo_conf]

Options:
`)
	flag.PrintDefaults()
}
