package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/*
	Input-levels
	- Start/Stop input-levels
	- Subscribe to input-levels
	- Unsubscribe from input-levels
	- Handle input-levels JSON data on callback
	- Save input-levels data in server properties
	- Forward input-levels JSON data to WebSocket clients
*/

type InputLevelsCommandRequest struct {
	Command    string   `json:"command"`
	Parameters []string `json:"parameters"`
}

// Subscribe to REW.app for "input-levels"
func (s *Server) inputLevelsCommand(command string) error {

	// Construct subscription request body
	subReq := InputLevelsCommandRequest{Command: command, Parameters: []string{}}
	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal inputLevelsCommand request: %v", err)
	}

	fmt.Printf("reqBody: %s\n", string(reqBody))

	// Make the POST request
	req, err := http.NewRequest("POST", s.rewEndpoint+"/input-levels/command", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post inputLevelsCommand request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("inputLevelsCommand request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("inputLevelsCommand request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

type InputLevelsSubscribeParameters struct {
	Unit string `json:"unit"`
}
type InputLevelsSubscribeRequest struct {
	Url        string                         `json:"url"`
	Parameters InputLevelsSubscribeParameters `json:"parameters"`
}

func (s *Server) inputLevelsSubscribe(url string, unit string) error {
	parameters := InputLevelsSubscribeParameters{Unit: unit}
	// Construct subscription request body
	subReq := InputLevelsSubscribeRequest{
		Url:        url,
		Parameters: parameters,
	}
	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal inputLevelsSubscribe request: %v", err)
	}

	// Make the POST request
	req, err := http.NewRequest("POST", s.rewEndpoint+"/input-levels/subscribe", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post inputLevelsSubscribe request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("inputLevelsSubscribe request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("inputLevelsSubscribe request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

type InputLevelsUnsubscribeRequest struct {
	Url        string                         `json:"url"`
	Parameters InputLevelsSubscribeParameters `json:"parameters"`
}

func (s *Server) inputLevelsUnsubscribe(url string, unit string) error {
	parameters := InputLevelsSubscribeParameters{Unit: unit}
	// Construct subscription request body
	subReq := InputLevelsSubscribeRequest{
		Url:        url,
		Parameters: parameters,
	}
	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal inputLevelsUnsubscribe request: %v", err)
	}

	// Make the POST request
	req, err := http.NewRequest("POST", s.rewEndpoint+"/input-levels/unsubscribe", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post inputLevelsUnsubscribe request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("inputLevelsUnsubscribe request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("inputLevelsUnsubscribe request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}

/*
	Start/Stop input-levels
*/

func (server *Server) startInputLevels(hook string) error {
	err := server.inputLevelsCommand("start")
	if err != nil {
		return err
	}

	err = server.inputLevelsSubscribe(hook, "dBFS")
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) stopInputLevels(hook string) error {
	err := server.inputLevelsUnsubscribe(hook, "dBFS")
	if err != nil {
		return err
	}

	err = server.inputLevelsCommand("stop")
	if err != nil {
		return err
	}

	return nil
}

/*
	Handle input-levels JSON data
*/

type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type InputLevelsSample struct {
	Unit            string    `json:"unit"`
	RMS             []float64 `json:"rms"`
	Peak            []float64 `json:"peak"`
	TimeSpanSeconds float64   `json:"timeSpanSeconds"`
}

// Handle HTTP POST requests and forward JSON to WebSocket clients
func (s *Server) handleDBFS(w http.ResponseWriter, r *http.Request) {
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

	// fmt.Printf("Received JSON: %s\n", body)

	// Verify if it's valid JSON
	sample := InputLevelsSample{}
	if err := json.Unmarshal(body, &sample); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(sample.RMS) > 0 {
		s.rewAPILeftdBFS = sample.RMS[0] // REW unit is configured as dBFS
		err = s.broadcast("Left_dBFS", sample.RMS[0])
		if err != nil {
			http.Error(w, "Failed to marshal metric JSON", http.StatusInternalServerError)
			return
		}
	}

	if len(sample.RMS) > 1 {
		s.rewAPIRightdBFS = sample.RMS[1] // REW Unit is configured as dBFS
		err = s.broadcast("Right_dBFS", sample.RMS[1])
		if err != nil {
			http.Error(w, "Failed to marshal metric JSON", http.StatusInternalServerError)
			return
		}
	}

}
