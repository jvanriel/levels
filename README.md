# levels

This code reads dBFS and dBSPL values of both left and right channels
of a miniDSP E.A.R.S microphone.

It compares these values with values read indirectly from REW using the REST API of REW.

Note that the signal sent to the headphones (mounted on the E.A.R.S is) is implemented
in a Web Audio application and outside of this code base. This browser application is 
playing a sine wave signal of a given frequency and gain. For testing the signal is 1 Hz 
and gain is max 1.0.

In the [results](/results/Screenshot.png) screenshot note that:

* MIC is the directly reading via Web Audio 
* CAL is the calibrated reading via REW

## Current Findings

* Both Java and Golang implementations give the same dBFS readings. 
  These readings are consistent with another implementation done with Web Audio in the browser.

* There is a difference of **3dB** between the dBFS readings done directly vs those done indirectly via REW.

* How to calculate dBSPL values exactly is not yet understood.

## Additional Questions

When calculating dBSPL from dBFS we should take into account:

* a fixed SPL offset (current default is 94)
* a variable offset according to the E.A.R.S calibration files
* a sensitivity factor found in the calibration files.

The go implementation has this function for it:

```
func (s *Server) adjust(channel int, dBFS float64) float64 {
	dBSPL := dBFS
	// Add fixed offset from options, default is 94.0
	dBSPL += float64(s.sploffset)

	// Add sensitivity from calibration files
	dBSPL += s.calfiles.sensitivity(channel)

	// Add interpolated SPL from calibration files
	dBSPL += s.calfiles.interpolatedSPL(channel)

	return dBSPL
}
```

It is yet unclear how to do this exacly:

* Why is there a 3dB difference in dBFS readings with REW? 
* What is the fixed SPL offset that we should use?. Current default is 94.
* What are we supposed to do with the sensitivity factor precisely?
* Do we just add the variable (frequency dependent) offsets?
* Is there other processing that should be done first such a filtering or weighting?


## Preparations

* Setup multi-channel REW for MacOS in its location and configure and test with E.A.R.S.
* Install golang (tested with go version go1.23.3 darwin/arm64) 
* Install Java (Tested with java 17 2021-09-14 LTS)

## Run Java code (from vscode)

* Select Levels.java file
* Right-click 

## Running go code in a separate terminal

run ```cd golang``` and ```go run```

Options are:

* with REW UI ```-withgui``` default is false (no REW UI, server only)
* frequency for calibration ```-frequency <value>``` default us 1000 (Hz)
* SPLOffset for dBSPL calculation from dBFS values ```-offset <value>``` default is 96 (dB) 


