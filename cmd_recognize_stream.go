package main

import (
	"bytes"
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/gordonklaus/portaudio"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

func infiniteSendBuffer(
	stream speechpb.Speech_StreamingRecognizeClient,
	reader io.Reader,
) {
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			err := stream.Send(&speechpb.StreamingRecognizeRequest{
				StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
					AudioContent: buf[:n],
				},
			})
			if err != nil {
				log.Printf("Could not send audio: %v", err)
			}
		}
		if err == io.EOF {
			// Nothing else to pipe, close the stream.
			if err := stream.CloseSend(); err != nil {
				log.Fatalf("Could not close stream: %v", err)
			}
			return
		}
		if err != nil {
			log.Printf("Could not read from stdin: %v", err)
			continue
		}
	}
}

func infiniteSendBufferByFile(
	stream speechpb.Speech_StreamingRecognizeClient,
	filename string,
) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	infiniteSendBuffer(stream, f)
}

func openAudioStream(in []int16) (*portaudio.Stream, error) {
	host, err := portaudio.DefaultHostApi()
	if err != nil {
		log.Fatal(err)
	}

	input := host.Devices[0].HostApi.DefaultInputDevice
	params := portaudio.LowLatencyParameters(input, nil)
	params.Input.Channels = 1
	params.Output.Channels = 0
	params.SampleRate = 16000
	params.FramesPerBuffer = len(in)
	audioStream, err := portaudio.OpenStream(params, in)
	return audioStream, err
}

func infiniteReadByAudioInput(
	buffer *bytes.Buffer,
	sig chan os.Signal,
	stream speechpb.Speech_StreamingRecognizeClient,
) {
	fatalIfError(portaudio.Initialize())
	time.Sleep(1)
	defer portaudio.Terminate()

	//in := make([]byte, 1024)
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

	fatalIfError(audioStream.Start())
loop:
	for {
		fatalIfError(audioStream.Read())
		fatalIfError(binary.Write(buffer, binary.LittleEndian, in))

		err = ioutil.WriteFile("/tmp/tmp.raw", buffer.Bytes(), 0644)
		fatalIfError(err)

		content, err := ioutil.ReadFile("/tmp/tmp.raw")
		fatalIfError(err)
		err = stream.Send(&speechpb.StreamingRecognizeRequest{
			StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
				AudioContent: content,
			},
		})
		if err != nil {
			log.Printf("Could not send audio: %v", err)
		}

		buffer.Reset()

		select {
		case <-sig:
			if err := stream.CloseSend(); err != nil {
				log.Fatalf("Could not close stream: %v", err)
			}
			break loop
		default:
		}
	}
}

func infiniteReceiveResult(stream speechpb.Speech_StreamingRecognizeClient) {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot stream results: %v", err)
		}
		if err := resp.Error; err != nil {
			// Workaround while the API doesn't give a more informative error.
			if err.Code == 3 || err.Code == 11 {
				log.Print("WARNING: Speech recognition request exceeded limit of 60 seconds.")
			}
			log.Fatalf("Could not recognize: %v", err)
		}
		for _, result := range resp.Results {
			fmt.Printf("Result: %+v\n", result)
		}
	}
	fmt.Println("infiniteReceiveResult finish")
}

func runStreamRecognize() {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	fatalIfError(err)

	stream, err := client.StreamingRecognize(ctx)
	fatalIfError(err)

	err = stream.Send(CreateStreamRecognizeConfigRequest())
	fatalIfError(err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	buf := make([]byte, 1024)
	buffer := bytes.NewBuffer(buf)

	go infiniteReadByAudioInput(buffer, sig, stream)
	//go infiniteSendBuffer(stream, buffer)

	//go infiniteSendBufferByFile(stream, "/tmp/b.raw")
	infiniteReceiveResult(stream)
}
