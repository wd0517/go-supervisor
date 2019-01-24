package main

import (
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func runAsDaemon(program func()) {
	daemon.AddCommand(daemon.StringFlag(&flagSignal, "shutdown"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(&flagSignal, "stop"), syscall.SIGTERM, termHandler)

	if !isDir(supervisor.config.LogPath) {
		os.MkdirAll(supervisor.config.LogPath, 0755)
	}

	logFile := filepath.Join(supervisor.config.LogPath, "supervisor.log")

	cntxt := &daemon.Context{
		PidFileName: supervisor.config.Pidfile,
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[go supervisord]"},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalln("Unable send signal to the daemon:", err)
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("<------------ Go-Supervisor started ----------->")

	go program()

	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("》============ Go-Supervisor terminated ==========《")
}

func termHandler(sig os.Signal) error {
	if sig == syscall.SIGQUIT {
		supervisor.stopAllProcesses()
	}
	return daemon.ErrStop
}
