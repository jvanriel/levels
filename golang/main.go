package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gordonklaus/portaudio"
)

/*
	Main
	- Start REW and start server
	- Subscribe to REW input-levels and SPL-meters
	- Start server
	- Wait (Use Ctrl-C to stop)
	- Unsubscribe from REW input-levels and SPL-meters
	- Stop REW
*/

func main() {
	// Define the -withgui flag
	withGUI := flag.Bool("withgui", false, "Start with GUI")
	frequency := flag.Int("frequency", 1000, "Frequency for SPL meter")
	calfiles := flag.String("calfiles", "ears", "Path to calibration files")
	sploffset := flag.Int("sploffset", 94, "Fixed SPL offset")

	// Parse the command-line flags
	flag.Parse()

	rewEndpoint := "http://localhost:4735"
	dBFSWebHook := "http://localhost:8080/dbfs"
	SPLWebHook := "http://localhost:8080/spl"
	c := make(chan os.Signal, 1)

	err := portaudio.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize PortAudio:", err)
	}

	calFiles := NewCalfiles(*calfiles, *frequency)
	err = calFiles.load()
	if err != nil {
		log.Fatalf("Error loading calibration files: %v", err)
	}

	server := NewServer(
		rewEndpoint,
		calFiles,
		*sploffset,
	)

	// Setup direct stream via portaudio

	stream, err := server.setupAudio()
	if err != nil {
		log.Fatalf("Failed to setup audio: %v", err)
	}

	err = stream.Start()
	if err != nil {
		log.Fatal("Failed to start PortAudio stream:", err)
	}

	// Handle WebSocket connections from browser and webhook callbacks from REW

	http.HandleFunc("/ws", server.handleWebSocket)
	http.HandleFunc("/dbfs", server.handleDBFS)
	http.HandleFunc("/spl", server.handleSPL)

	// Start the server with error handling for port conflict

	proc, err := server.startREW(rewEndpoint, *withGUI)
	if err != nil {
		log.Fatalf("Failed to start REW: %v", err)
	}

	// Subscribe to REW input-levels and SPL-meters

	err = server.rewSelectInputDevice("E.A.R.S Gain: 18dB")
	if err != nil {
		log.Printf("Failed to select input device: %v\n", err)
		goto process_stop
	}

	err = server.startInputLevels(dBFSWebHook)
	if err != nil {
		log.Printf("Failed to start input-levels: %v\n", err)
		goto process_stop
	}

	err = server.startSPLMeters(SPLWebHook)
	if err != nil {
		log.Printf("Failed to start spl-meters: %v\n", err)
		goto spl_meter_unsubscribe
	}

	// Show last levels
	go func() {
		for {
			fmt.Printf("Direct Left: %7.2f dBFS %7.2f dBSPL - Right: %7.2f dBFS %7.2f dBSPL\n",
				server.directLeftdBFS, server.directLeftdBSPL,
				server.directRightdBFS, server.directRightdBSPL,
			)
			fmt.Printf("REWAPI Left: %7.2f dBFS %7.2f dBSPL - Right: %7.2f dBFS %7.2f dBSPL\n",
				server.rewAPILeftdBFS, server.rewAPILeftdBSPL,
				server.rewAPIRightdBFS, server.rewAPIRightdBSPL,
			)
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	// Start server in go routine
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	// Wait for Ctrl-C

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

spl_meter_unsubscribe:

	err = server.stopSPLMeters(SPLWebHook)
	if err != nil {
		log.Printf("Failed to unsubscribe to spl-meter: %v\n", err)
	}

	err = server.stopInputLevels(dBFSWebHook)
	if err != nil {
		log.Printf("Failed to stop input-levels: %v", err)
	}

process_stop:

	server.stopREW(proc)

	log.Println("Server stopped")
}
