package main

import (
	"errors"
)

type Supervisor struct {
	config *SupervisorConfig
	Pros   map[string]*SubProcess
}

func (s *Supervisor) runforever() {
	s.startAllProcesses(true)
}

func (s *Supervisor) startProcess(name string) error {
	process, ok := s.Pros[name]
	if !ok {
		return errors.New("Process not exists")
	}
	if process.isRunning() {
		return errors.New("Process already running")
	}
	process.spawn()
	return nil
}

func (s *Supervisor) stopProcess(name string) error {
	process, ok := s.Pros[name]
	if !ok {
		return errors.New("Process not exists")
	}
	if !process.isRunning() {
		return errors.New("Process not running")
	}
	process.stop()
	return nil
}

func (s *Supervisor) restartProcess(name string) error {
	process, ok := s.Pros[name]
	if !ok {
		return errors.New("Process not exists")
	}
	if !process.isRunning() {
		return errors.New("Process not running")
	}
	process.stop()
	process.spawn()
	return nil
}

func (s *Supervisor) startAllProcesses(init bool) {
	for _, process := range s.Pros {
		if !process.isRunning() && (!init || process.config.AutoStart) {
			process.spawn()
		}
	}
}

func (s *Supervisor) stopAllProcesses() {
	for _, process := range s.Pros {
		if process.isRunning() {
			process.stop()
		}
	}
}
