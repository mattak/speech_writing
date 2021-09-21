package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	RecordSplitCmd = &cobra.Command{
		Use:     "record",
		Short:   "record audio into raw files",
		Long:    "record audio into raw files",
		Example: " speech_writing record /tmp/records/rec_%04d.raw 2000 16000",
		Run:     runCommandRecordSplit,
	}
)

type RecordingState struct {
	SamplingRate      int // 16000
	ActiveThreshold   int16
	ActiveEndDuration int

	IsActive              bool
	IsActiveStarting      bool
	IsInactiveStarting    bool
	ActiveStartingIndex   int
	InactiveStartingIndex int

	OutputFileFormat string
	ChunkNumber      int
	BufferOffset     int
	Buffer           []int16
}

func runCommandRecordSplit(cmd *cobra.Command, args []string) {
	if len(args) < 3 {
		log.Fatal("usage: <record_path_format> <active_threshold> <active_end_duration>")
	}
	recordPathFormat := args[0]
	activeThreshold, err := strconv.ParseInt(args[1], 10, 64)
	fatalIfError(err)
	activeEndDuration, err := strconv.ParseInt(args[2], 10, 64)
	fatalIfError(err)

	RunRecordSplit(recordPathFormat, int16(activeThreshold), int(activeEndDuration))
}

func (state *RecordingState) Process(data []int16, isLastBuffer bool) {
	state.BufferOffset = len(state.Buffer)
	state.Buffer = append(state.Buffer, data...)

	for i := state.BufferOffset; i < len(state.Buffer); i++ {
		if !state.IsActive {
			if state.Buffer[i] >= state.ActiveThreshold || state.Buffer[i] <= -state.ActiveThreshold {
				fmt.Fprintln(os.Stderr, "Record", state.ChunkNumber)
				state.IsActive = true
				state.IsInactiveStarting = false
				state.ActiveStartingIndex = i
			}
			continue
		}

		// inactive
		if !state.IsInactiveStarting {
			if state.Buffer[i] < state.ActiveThreshold && state.Buffer[i] > -state.ActiveThreshold {
				state.IsInactiveStarting = true
				state.InactiveStartingIndex = i
			} else {
				state.IsInactiveStarting = false
			}
			continue
		}

		if state.Buffer[i] >= state.ActiveThreshold || state.Buffer[i] <= -state.ActiveThreshold {
			state.IsInactiveStarting = false
			continue
		}

		if i-state.InactiveStartingIndex > state.ActiveEndDuration || (isLastBuffer && i >= len(state.Buffer)-1) {
			recordSeconds := float64(i-state.ActiveStartingIndex) / float64(state.SamplingRate)
			state.IsActive = false
			state.IsInactiveStarting = false
			state.InactiveStartingIndex = 0

			fileName := fmt.Sprintf(state.OutputFileFormat, state.ChunkNumber)
			fmt.Fprintln(os.Stderr, "write", fileName, recordSeconds)
			WriteAudioRawFile(fileName, state.Buffer, state.ActiveStartingIndex, i)
			state.ChunkNumber++
		}
	}

	if state.IsActive {
		// active index 以前をcutoff
		state.Buffer = state.Buffer[state.ActiveStartingIndex:]
		state.BufferOffset = len(state.Buffer)
		if state.IsInactiveStarting {
			if state.InactiveStartingIndex > state.ActiveStartingIndex {
				state.InactiveStartingIndex = state.InactiveStartingIndex - state.ActiveStartingIndex
			} else {
				state.InactiveStartingIndex = 0
			}
		}
		state.ActiveStartingIndex = 0
	} else {
		// inactiveなら全部捨てる
		state.Buffer = []int16{}
		state.BufferOffset = 0
	}
}

func RunRecordSplit(
	outputFilePathFormat string,
	activeThreshold int16,
	activeEndDuration int,
) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// init
	fatalIfError(portaudio.Initialize())
	time.Sleep(1)
	defer portaudio.Terminate()

	// init record
	recordingState := RecordingState{
		SamplingRate:          16000,
		ActiveThreshold:       activeThreshold,
		ActiveEndDuration:     activeEndDuration,
		IsActive:              false,
		IsActiveStarting:      false,
		IsInactiveStarting:    false,
		ActiveStartingIndex:   0,
		InactiveStartingIndex: 0,

		OutputFileFormat: outputFilePathFormat,
		ChunkNumber:      0,
		BufferOffset:     0,
		Buffer:           []int16{},
	}

	// open audio
	in := make([]int16, 512)
	audioStream, err := portaudio.OpenDefaultStream(
		1,
		0,
		16000,
		len(in),
		in,
	)
	fatalIfError(err)
	defer audioStream.Close()

	// start audio
	fatalIfError(audioStream.Start())
loop:
	for {
		fatalIfError(audioStream.Read())
		recordingState.Process(in, false)

		select {
		case <-sig:
			fatalIfError(audioStream.Read())
			recordingState.Process(in, true)
			break loop
		default:
		}
	}
}
