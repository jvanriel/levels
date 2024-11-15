package main

import (
	"fmt"
	"math"

	"github.com/gordonklaus/portaudio"
)

/*
	Multichannel audio input
	- Setup portaudio
	- Open a stream for "E.A.R.S Gain: 18dB"
	- Read audio samples from the stream
	- Separate audio samples into left and right channels
	- Calculate RMS for each channel
	- Calculate dBFS for each channel
	- Calculate dBSPL for each channel
	- Save the last calculated values in server properties
*/

func (s *Server) setupAudio() (*portaudio.Stream, error) {

	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}
	defer portaudio.Terminate()

	apis, err := portaudio.HostApis()
	if err != nil {
		return nil, err
	}

	for i, api := range apis {
		fmt.Printf("Host API %d: %s\n", i, api.Name)
	}

	var inDev *portaudio.DeviceInfo

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	name := "E.A.R.S Gain: 18dB"

	for i, dev := range devices {
		fmt.Printf("Device %d: %s\n", i, dev.Name)
		if dev.Name == name {
			inDev = devices[i]
		}
	}

	if inDev == nil {
		return nil, fmt.Errorf("input device '%s' not found", name)
	}

	p := portaudio.HighLatencyParameters(inDev, nil)
	p.Input.Channels = 2
	p.Output.Channels = 0
	p.SampleRate = 48000
	p.FramesPerBuffer = 2048

	stream, err := portaudio.OpenStream(p, s.readAudio)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *Server) readAudio(in []float32) {
	// Separate audio samples into left and right channels and calculate RMS for each
	// We assume the audio buffer in is interleaved (i.e., [left, right, left, right, ...]).
	// This is typical for stereo audio data but should be confirmed with our specific setup.

	var sumSquaresLeft, sumSquaresRight float64
	numSamples := len(in) / 2 // Number of frames (each frame has two samples: left and right)

	for i := 0; i < len(in); i += 2 {
		leftSample := in[i]
		rightSample := in[i+1]

		sumSquaresLeft += float64(leftSample * leftSample)
		sumSquaresRight += float64(rightSample * rightSample)
	}

	// Calculate RMS for each channel
	rmsLeft := math.Sqrt(sumSquaresLeft / float64(numSamples))
	rmsRight := math.Sqrt(sumSquaresRight / float64(numSamples))

	// Calculate SPL for each channel in dB SPL (using a reference RMS level of 1.0)
	s.directLeftdBFS = 20 * math.Log10(rmsLeft)
	s.directRightdBFS = 20 * math.Log10(rmsRight)
	s.directLeftdBSPL = s.adjust(0, s.directLeftdBFS)
	s.directRightdBSPL = s.adjust(1, s.directRightdBFS)

	s.counter++
}
