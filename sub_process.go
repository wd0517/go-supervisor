package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
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

	stdoutLogger *lumberjack.Logger
	stderrLogger *lumberjack.Logger

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
		log.Printf("Process: [%s] start failed. %s", sp.Name, err.Error())
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
	sp.stdoutLogger.Close()
	sp.stderrLogger.Close()

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
		log.Printf("Process: [%s] relive, backoff: %d / %d", sp.Name, sp.backoff+1, sp.config.StartRetries)
		sp.spawn()
	} else {
		sp.State = FATAL
		log.Printf("Process: [%s] fall into fatal error, give up!", sp.Name)
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

	outFileName := filepath.Join(logDir, "stdout.log")
	errFileName := filepath.Join(logDir, "stderr.log")

	sp.stdoutLogger = &lumberjack.Logger{
		Filename:   outFileName,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true,
	}

	sp.stderrLogger = &lumberjack.Logger{
		Filename:   errFileName,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true,
	}

	cmd.Stdout = sp.stdoutLogger
	cmd.Stderr = sp.stderrLogger

	if sp.config.Directory != "" {
		cmd.Dir = sp.config.Directory
	}

	return cmd
}
