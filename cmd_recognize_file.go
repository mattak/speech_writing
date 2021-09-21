package main

import (
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"fmt"
	"io/ioutil"
)

func RunFileRecognize(filePaths []string) {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	fatalIfError(err)

	for _, path := range filePaths {
		content, err := ioutil.ReadFile(path)
		fatalIfError(err)
		response, err := client.Recognize(ctx, CreateFileRecognizeConfigRequest(content))
		fatalIfError(err)

		for _, result := range response.Results {
			for _, alt := range result.Alternatives {
				fmt.Println(alt.Transcript)
			}
		}
	}
}

