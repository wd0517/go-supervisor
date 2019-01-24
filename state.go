package main

type ProcessState string

const (
	STOPPED  = ProcessState("STOPPED")
	RUNNING  = ProcessState("RUNNING")
	BACKOFF  = ProcessState("BACKOFF")
	STOPPING = ProcessState("STOPPING")
	FATAL    = ProcessState("FATAL")
)

var (
	STOPPED_STATES = []ProcessState{STOPPED, FATAL}
	RUNNING_STATES = []ProcessState{RUNNING, BACKOFF}
)
