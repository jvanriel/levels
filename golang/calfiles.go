package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type DataPoint struct {
	Frequency float64
	SPL       float64
	Phase     float64
}

type CalFiles struct {
	frequency        float64
	folder           string
	leftDataPoints   []DataPoint
	leftSensitivity  float64
	rightDataPoints  []DataPoint
	rightSensitivity float64
}

func NewCalfiles(folder string, frequency int) *CalFiles {
	return &CalFiles{
		frequency: float64(frequency),
		folder:    folder,
	}
}

func (c *CalFiles) load() error {
	// Load calibration files for left and right channels

	// Open the directory
	dir, err := os.Open(c.folder)
	if err != nil {
		return fmt.Errorf("error opening folder: %w", err)
	}
	defer dir.Close()

	// Read the directory contents
	files, err := dir.Readdir(-1) // -1 means read all files and folders
	if err != nil {
		return fmt.Errorf("error reading folder contents: %w", err)
	}

	// Iterate over the files and folders
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".txt") {
				if err := c.loadFile(file); err != nil {
					return fmt.Errorf("error loading calibration file: %v", err)
				}
			}
		}
	}

	if len(c.leftDataPoints) == 0 {
		return fmt.Errorf("no calibration data found for LEFT channel")
	}
	if len(c.rightDataPoints) == 0 {
		return fmt.Errorf("no calibration data found for RIGHT channel")
	}

	return nil
}

func (c *CalFiles) loadFile(fileInfo os.FileInfo) error {
	// Open the calibration file
	file, err := os.Open(c.folder + "/" + fileInfo.Name())
	if err != nil {
		return fmt.Errorf("error opening calibration file %s: %v", fileInfo.Name(), err)
	}
	defer file.Close()

	sensitiviy := 0.0
	channel := -1 // 0 for LEFT, 1 for RIGHT, -1 for unknown

	// Read the calibration data line by line
	var data []DataPoint
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header and metadata lines starting with "*" or other non-data lines
		if strings.HasPrefix(line, "*") || strings.HasPrefix(line, "\"") || len(strings.TrimSpace(line)) == 0 {
			if strings.Contains(line, "LEFT") {
				channel = 0
			}
			if strings.Contains(line, "RIGHT") {
				channel = 1
			}
			if strings.Contains(line, "Sens Factor") {
				parts := strings.Split(line, ",")
				if len(parts) > 0 {
					factorField := strings.TrimSpace(parts[0])
					p := strings.Split(factorField, "=")
					if len(p) != 2 {
						return fmt.Errorf("invalid sensitivity data parsing factor field: %s", line)
					}
					numberOnly := strings.TrimSuffix(p[1], "dB")
					sens, err := strconv.ParseFloat(numberOnly, 64)
					if err != nil {
						log.Fatalf("invalid sensitivity data parsing float '%s' %v", p[1], err)
					}
					sensitiviy = sens
				} else {
					return fmt.Errorf("invalid sensitivity data parsing sense factor line %s", line)
				}
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			return fmt.Errorf("invalid calibration data: %s", line)
		}
		frequency, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return fmt.Errorf("error parsing frequency: %v", err)
		}
		spl, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return fmt.Errorf("error parsing SPL: %v", err)
		}
		phase, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return fmt.Errorf("error parsing phase: %v", err)
		}
		data = append(data, DataPoint{Frequency: frequency, SPL: spl, Phase: phase})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading calibration file: %v", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("no calibration data found in file: %s", fileInfo.Name())
	}

	if channel == -1 {
		return fmt.Errorf("no channel found in calibration file: %s", fileInfo.Name())
	}

	if channel == 0 {
		c.leftSensitivity = sensitiviy
		c.leftDataPoints = data
	} else if channel == 1 {
		c.rightSensitivity = sensitiviy
		c.rightDataPoints = data
	} else {
		return fmt.Errorf("unknown channel in calibration file: %s", fileInfo.Name())
	}
	return nil
}

// InterpolateSPL takes a frequency and returns the interpolated SPL value.
func interpolateSPL(frequency float64, data []DataPoint) (float64, error) {
	if frequency < data[0].Frequency || frequency > data[len(data)-1].Frequency {
		return 0, fmt.Errorf("frequency %.2f is out of range", frequency)
	}

	// Search for the two data points surrounding the frequency
	for i := 0; i < len(data)-1; i++ {
		if frequency >= data[i].Frequency && frequency <= data[i+1].Frequency {
			// Linear interpolation formula
			x0 := data[i].Frequency
			x1 := data[i+1].Frequency
			y0 := data[i].SPL
			y1 := data[i+1].SPL

			// Interpolated SPL value
			spl := y0 + (y1-y0)*(frequency-x0)/(x1-x0)
			return spl, nil
		}
	}

	return 0, fmt.Errorf("frequency %.2f not found", frequency)
}

func (c *CalFiles) interpolatedSPL(channel int) float64 {
	if channel == 0 {
		spl, err := interpolateSPL(c.frequency, c.leftDataPoints)
		if err != nil {
			log.Printf("Error interpolating SPL: %v", err)
			return 0.0
		}
		return spl
	} else {
		spl, err := interpolateSPL(c.frequency, c.rightDataPoints)
		if err != nil {
			log.Printf("Error interpolating SPL: %v", err)
		}
		return spl
	}
}

func (c *CalFiles) sensitivity(channel int) float64 {
	if channel == 0 {
		return c.leftSensitivity
	} else {
		return c.rightSensitivity
	}
}
