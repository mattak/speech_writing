package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func WriteAudioRawFile(
	fileName string,
	data []int16,
	startIndex int,
	endIndex int,
) {
	tmpFileName := fmt.Sprintf("%s.tmp", fileName)
	f, err := os.Create(tmpFileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	err = binary.Write(f, binary.LittleEndian, data[startIndex:endIndex])
	if err != nil {
		log.Fatalln(err)
	}

	err = os.Rename(tmpFileName, fileName)
	fatalIfError(err)
}
