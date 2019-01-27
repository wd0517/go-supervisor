package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type ResJson struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func startWebServer(listenAddr string) {
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/start/", actionHandler)
	http.HandleFunc("/stop/", actionHandler)
	http.HandleFunc("/restart/", actionHandler)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Printf("Failed to start web server, %s", err.Error())
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s: %s %q", r.RemoteAddr, r.Method, r.URL)
	js, err := json.Marshal(supervisor.Pros)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func actionHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s: %s %q", r.RemoteAddr, r.Method, r.URL)

	var rJson = ResJson{}
	var query = r.URL.Query()
	var err error

	if name, ok := query["name"]; !ok {
		rJson.Code = 0
		rJson.Message = "?name is required"
	} else {
		switch r.URL.Path {
		case "/start/":
			err = supervisor.startProcess(name[0])
		case "/stop/":
			err = supervisor.stopProcess(name[0])
		case "/restart/":
			err = supervisor.restartProcess(name[0])
		}
		if err != nil {
			rJson.Code = 1
			rJson.Message = err.Error()
		} else {
			rJson.Code = 200
			rJson.Message = "success"
		}
	}

	js, _ := json.Marshal(rJson)
	w.Write(js)
	w.Header().Set("Content-Type", "application/json")
}
