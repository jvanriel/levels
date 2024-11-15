package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

/*
	Server
	- Handle WebSocket connections
	- Broadcast data to WebSocket clients
	- Start/Stop REW
	- Select audio input device in REW
*/

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo purposes
	},
}

type Server struct {
	clients     map[*websocket.Conn]bool
	mu          sync.Mutex
	rewEndpoint string
	sploffset   int
	calfiles    *CalFiles

	rewAPILeftdBFS   float64
	rewAPIRightdBFS  float64
	rewAPILeftdBSPL  float64
	rewAPIRightdBSPL float64

	directLeftdBFS   float64
	directRightdBFS  float64
	directLeftdBSPL  float64
	directRightdBSPL float64

	counter int
}

func NewServer(rewEndpoint string, calFiles *CalFiles, sploffset int) *Server {
	var server = &Server{
		rewEndpoint: rewEndpoint,
		clients:     make(map[*websocket.Conn]bool),
		sploffset:   sploffset,
		calfiles:    calFiles,
	}
	return server
}

// Handle WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	log.Println("New WebSocket client connected")

	// Keep the connection open until the client disconnects
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)
			break
		}
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
	log.Println("WebSocket client disconnected")
}

func (s *Server) broadcast(name string, value float64) error {

	metric := Metric{
		Name:  name,
		Value: value,
	}

	body, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	// Broadcast the JSON to all connected WebSocket clients
	s.mu.Lock()
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, body)
		if err != nil {
			log.Println("Error sending message:", err)
			client.Close()
			delete(s.clients, client)
		}
	}
	s.counter++
	s.mu.Unlock()

	return nil
}

func (s *Server) startREW(url string, withgui bool) (*os.Process, error) {

	path := "/Applications/REW/REW.app/Contents/MacOS/JavaApplicationStub"

	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}

	args := []string{"REW.app", "-api"}
	if !withgui {
		args = append(args, "-nogui")
	}

	attr := &os.ProcAttr{
		Dir:   "",
		Env:   nil,
		Files: []*os.File{os.Stdin, devnull, os.Stderr}, // process detachted from stdout
		Sys: &syscall.SysProcAttr{
			Setpgid: true, // This places the child in a new process group
		},
	}

	proc, err := os.StartProcess(path, args, attr)
	if err != nil {
		return nil, err
	}

	go func() {
		_, err = proc.Wait()
		if (err != nil) && (err.Error() != "signal: killed") {
			fmt.Printf("Command finished with error: %v\n", err)
		}
		devnull.Close()
	}()

	fmt.Println("REW started pid:", proc.Pid)

	// Need to give it some time to settle
	time.Sleep(10000 * time.Millisecond)

	// Wait for REW http to start
	for {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Need to give it some time to settle
	time.Sleep(3000 * time.Millisecond)

	fmt.Println("REW ready on:", url)

	return proc, nil
}

func (s *Server) stopREW(proc *os.Process) error {
	fmt.Println("Shutting down...", proc.Pid)
	return proc.Signal(syscall.SIGKILL)
}

type AudioSelectInputDeviceRequest struct {
	Device string `json:"device"`
}

func (s *Server) rewSelectInputDevice(device string) error {
	subReq := AudioSelectInputDeviceRequest{
		Device: device,
	}

	reqBody, err := json.Marshal(subReq)
	if err != nil {
		return fmt.Errorf("failed to marshal AudioSelectInputDeviceRequest request: %v", err)
	}

	fmt.Printf("rewEndpoint: %s\n", string(s.rewEndpoint))
	fmt.Println(string(reqBody))

	req, err := http.NewRequest("POST", s.rewEndpoint+"/audio/java/input-device", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to post AudioSelectInputDeviceRequest request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("AudioSelectInputDeviceRequest request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AudioSelectInputDeviceRequest request failed with status %d: %s", resp.StatusCode, string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(body))
	}

	return nil
}
