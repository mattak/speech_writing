package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// fileName: /tmp/rec.raw
// outputFileFormat: /tmp/rec_%04d.raw
// activeThreshold: 2000
// activeEndDuration: 16000
func RunSplit(
	fileName string,
	outputFileFormat string,
	activeThreshold int16,
	activeEndDuration int,
) {
	f, err := os.Open(fileName)
	fatalIfError(err)
	defer f.Close()

	fileInfo, err := f.Stat()
	fatalIfError(err)
	dataSize := int(fileInfo.Size() / 2)
	data := make([]int16, dataSize)

	err = binary.Read(f, binary.LittleEndian, data)
	fatalIfError(err)

	isActive := false
	isInactiveStarting := false
	activeStartingIndex := 0
	inactiveStartingIndex := 0
	chunkNumber := 1
	for i := 0; i < dataSize; i++ {
		if !isActive {
			if data[i] >= activeThreshold || data[i] <= -activeThreshold {
				fmt.Println("active", float64(i)/16000.0)
				isActive = true
				isInactiveStarting = false
				activeStartingIndex = i
			}
			continue
		}

		// inactive
		if !isInactiveStarting {
			if data[i] < activeThreshold && data[i] > -activeThreshold {
				isInactiveStarting = true
				inactiveStartingIndex = i
			} else {
				isInactiveStarting = false
			}
			continue
		}

		if data[i] >= activeThreshold || data[i] <= -activeThreshold {
			isInactiveStarting = false
			continue
		}

		if i-inactiveStartingIndex > activeEndDuration || i >= dataSize-1 {
			fmt.Println("inactive", float64(i)/16000.0)
			isActive = false
			isInactiveStarting = false
			inactiveStartingIndex = 0

			fileName := fmt.Sprintf(outputFileFormat, chunkNumber)
			WriteAudioRawFile(fileName, data, activeStartingIndex, i)
			chunkNumber++
		}
	}
}
