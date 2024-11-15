package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

/*
	SPL Meter
	- Configure SPL Meter
	- Subscribe to SPL Meter
	- Unsubscribe from SPL Meter
	- Handle SPL Meter JSON data on callback
	- Save SPL Meter data in server properties
	- Forward SPL Meter JSON data to WebSocket clients
*/

type SPLMeterConfiguration struct {
	Mode              string `json:"mode"`
	Weighting         string `json:"weighting"`
	Filter            string `json:"filter"`
	HighPassActive    bool   `json:"highPassActive"`
	RollingLeqActive  bool   `json:"rollingLeqActive"`
	RollingLeqMinutes int    `json:"rollingLeqMinutes"`
}

func (s *Server) splMeterConfigure(meter int) error {
	cfgReq := SPLMeterConfiguration{
		Mode:              "SPL",
		Weighting:         "Z",
		Filter:            "Fast",
		HighPassActive:    true,
		RollingLeqActive:  true,
		RollingLeqMinutes: 1,
	}

	reqBody, err := json.Marshal(cfgReq)
	if err != nil {
		return fmt.Errorf("failed to marshal SPLMeterConfiguration: %v", err)
	}

	req, err := http.NewRequest("POST", s.rewEndpoint+"/spl-meter/"+strconv.Itoa(meter)+"/configuration", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post SPLMeterConfiguration request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SPLMeterConfiguration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SPLMeterConfiguration request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

type SPLMeterSubscribeParameters struct {
}

type SPLMeterSubscribeRequest struct {
	Url        string                      `json:"url"`
	Parameters SPLMeterSubscribeParameters `json:"parameters"`
}

func (s *Server) splMeterSubscribe(meter int, url string) error {
	parameters := SPLMeterSubscribeParameters{}

	subReq := SPLMeterSubscribeRequest{
		Url:        url,
		Parameters: parameters,
	}

	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal SPLMeterSubscribeRequest request: %v", err)
	}

	req, err := http.NewRequest("POST", s.rewEndpoint+"/spl-meter/"+strconv.Itoa(meter)+"/subscribe", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post SPLMeterSubscribeRequest request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SPLMeterSubscribeRequest request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SPLMeterSubscribeRequest request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

type SPLMeterUnsubscribeParameters struct {
}

type SPLMeterUnsubscribeRequest struct {
	Url        string                        `json:"url"`
	Parameters SPLMeterUnsubscribeParameters `json:"parameters"`
}

func (s *Server) splMeterUnsubscribe(meter int, url string) error {
	parameters := SPLMeterUnsubscribeParameters{}

	subReq := SPLMeterUnsubscribeRequest{
		Url:        url,
		Parameters: parameters,
	}
	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal SPLMeterUnsubscribeRequest request: %v", err)
	}

	req, err := http.NewRequest("POST", s.rewEndpoint+"/spl-meter/"+strconv.Itoa(meter)+"/unsubscribe", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post SPLMeterUnsubscribeRequest request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SPLMeterUnsubscribeRequest request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SPLMeterSubscribeRequest request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

type SPLMeterCommandRequest struct {
	Command    string   `json:"command"`
	Parameters []string `json:"parameters"`
}

func (s *Server) splMeterCommand(meter int, command string) error {
	parameters := []string{}

	subReq := SPLMeterCommandRequest{
		Command:    command,
		Parameters: parameters,
	}

	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal SPLMeterCommandRequest request: %v", err)
	}

	req, err := http.NewRequest("POST", s.rewEndpoint+"/spl-meter/"+strconv.Itoa(meter)+"/command", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post SPLMeterCommandRequest request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SPLMeterCommandRequest request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SPLMeterCommandRequest request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

/*
	Http handler for SPLMeter samples
*/

type SPLMeterSample struct {
	MeterNumber       int     `json:"meterNumber"`
	Weighting         string  `json:"weighting"`
	Filter            string  `json:"filter"`
	SPL               float64 `json:"spl"`
	Leq               float64 `json:"leq"`
	IsRollingLeq      bool    `json:"isRollingLeq"`
	RollingLeqMinutes float64 `json:"rollingLeqMinutes"`
	Leq1m             float64 `json:"leq1m"`
	Leq10m            float64 `json:"leq10m"`
	Sel               float64 `json:"sel"`
	ElapsedTime       float64 `json:"elapsedTime"`
}

/*
	Adjust SPL Meter samples
*/

// Handle HTTP POST requests and forward JSON to WebSocket clients
func (s *Server) handleSPL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the JSON body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify if it's valid JSON
	sample := SPLMeterSample{}
	if err := json.Unmarshal(body, &sample); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	label := ""
	if sample.MeterNumber == 1 {
		label = "Left_dBSPL"
		s.rewAPILeftdBSPL = sample.SPL
		if err := s.broadcast("Left_dBSPL", sample.SPL); err != nil {
			http.Error(w, "Failed to broadcast Left_dBSPL", http.StatusInternalServerError)
			return
		}
	} else {
		label = "Right_dBSPL"
		s.rewAPIRightdBSPL = sample.SPL
		if err := s.broadcast("Right_dBSPL", sample.SPL); err != nil {
			http.Error(w, "Failed to marshal metric JSON", http.StatusInternalServerError)
			return
		}
	}

	metric := Metric{
		Name:  label,
		Value: sample.SPL,
	}

	metricBody, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, "Failed to marshal metric JSON", http.StatusInternalServerError)
		return
	}

	// Broadcast the JSON to all connected WebSocket clients
	s.mu.Lock()
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, metricBody)
		if err != nil {
			log.Println("Error sending message:", err)
			client.Close()
			delete(s.clients, client)
		}
	}
	s.mu.Unlock()

}

/*
	Start/Stop spl-meters
*/

func (server *Server) startSPLMeter(meter int, hook string) error {

	err := server.splMeterConfigure(meter)
	if err != nil {
		return err
	}

	err = server.splMeterCommand(meter, "start")
	if err != nil {
		return err
	}

	err = server.splMeterSubscribe(meter, hook)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) startSPLMeters(hook string) error {
	err := server.startSPLMeter(1, hook)
	if err != nil {
		return err
	}

	err = server.startSPLMeter(2, hook)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) stopSPLMeter(meter int, hook string) error {

	err := server.splMeterUnsubscribe(meter, hook)
	if err != nil {
		return err
	}

	err = server.splMeterCommand(meter, "stop")
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) stopSPLMeters(hook string) error {
	err := server.stopSPLMeter(1, hook)
	if err != nil {
		return err
	}

	err = server.stopSPLMeter(2, hook)
	if err != nil {
		return err
	}

	return nil
}
