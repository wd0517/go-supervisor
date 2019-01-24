package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type SubProcess struct {
	Pid   int
	Name  string
	State ProcessState

	config *ProcessConfig

	adminStop bool
	backoff   int
	lastStart time.Time
	lastStop  time.Time
	outFile   *os.File
	errFile   *os.File

	cmd *exec.Cmd
	mu  sync.Mutex
}

func (sp *SubProcess) isRunning() bool {
	for _, state := range RUNNING_STATES {
		if sp.State == state {
			return true
		}
	}
	return false
}

func (sp *SubProcess) spawn() {
	if sp.Pid != 0 {
		return
	}

	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.lastStart = time.Now()
	sp.adminStop = false
	sp.cmd = sp.buildCmd()

	if err := sp.cmd.Start(); err != nil {
		log.Printf("Process: [%s] start failed", sp.Name)
	} else {
		log.Printf("Process: [%s] start successful", sp.Name)
		sp.Pid = sp.cmd.Process.Pid
		sp.State = RUNNING
	}

	go sp.transition()
}

func (sp *SubProcess) transition() {
	if sp.Pid != 0 {
		sp.cmd.Wait()
	}

	sp.Pid = 0
	sp.lastStop = time.Now()
	sp.outFile.Close()
	sp.errFile.Close()

	if sp.adminStop {
		sp.State = STOPPED
		log.Printf("Process: [%s] stop successful", sp.Name)
		return
	} else {
		sp.State = BACKOFF
		log.Printf("Process: [%s] backoff", sp.Name)
	}

	if time.Now().Sub(sp.lastStart).Seconds() <= sp.config.StartSecs {
		sp.backoff += 1
	} else {
		sp.backoff = 0
	}

	if sp.backoff < sp.config.StartRetries && sp.config.AutoreReStart {
		log.Printf("Process: [%s] relive", sp.Name)
		sp.spawn()
	} else {
		sp.State = FATAL
	}
}

func (sp *SubProcess) stop() {
	if sp.Pid == 0 {
		return
	}

	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.adminStop = true
	sp.State = STOPPING
	sp.backoff = 0
	sp.cmd.Process.Signal(syscall.SIGTERM)

	select {
	case <-time.After(time.Duration(2) * time.Second):
		if sp.State != STOPPED {
			log.Printf("Process: [%s] force kill", sp.Name)
			sp.cmd.Process.Signal(syscall.SIGKILL)
		}
	}
}

func (sp *SubProcess) buildCmd() *exec.Cmd {
	cmd := exec.Command("/bin/bash", "-c", sp.config.Command)

	logDir := filepath.Join(supervisor.config.LogPath, sp.Name)
	if !isDir(logDir) {
		os.MkdirAll(logDir, 0755)
	}

	outFile := filepath.Join(logDir, "stdout.log")
	errFile := filepath.Join(logDir, "stderr.log")

	sp.outFile, _ = os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	sp.errFile, _ = os.OpenFile(errFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	cmd.Stderr = io.MultiWriter(sp.errFile)
	cmd.Stdout = io.MultiWriter(sp.outFile)

	if sp.config.Directory != "" {
		cmd.Dir = sp.config.Directory
	}

	return cmd
}
