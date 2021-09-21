package main

import (
	"encoding/binary"
	"github.com/gordonklaus/portaudio"
	"os"
	"os/signal"
	"time"
)

func RunRecord(filePath string) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// init
	fatalIfError(portaudio.Initialize())
	time.Sleep(1)
	defer portaudio.Terminate()

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

	// open file
	f, err := os.Create(filePath)
	fatalIfError(err)
	defer f.Close()

	// start audio
	fatalIfError(audioStream.Start())
loop:
	for {
		fatalIfError(audioStream.Read())
		fatalIfError(binary.Write(f, binary.LittleEndian, in))

		select {
		case <-sig:
			break loop
		default:
		}
	}
}
