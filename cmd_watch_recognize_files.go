package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var (
	WatchRecognizeFilesCmd = &cobra.Command{
		Use:     "watch_recognize",
		Short:   "recognize audio raw files by directory watching",
		Long:    "recognize audio raw files by directory watching",
		Example: " speech_writing watch_recognize /tmp/records/",
		Run:     runCommandWatchRecognizeFiles,
	}
)

func runCommandWatchRecognizeFiles(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("usage: <watch_directory>")
	}
	dir := args[0]
	RunWatchRecognize(dir)
}

func RunWatchRecognize(dir string) {
	watcher, err := fsnotify.NewWatcher()
	fatalIfError(err)
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if strings.HasSuffix(event.Name, ".raw") {
						fmt.Fprintln(os.Stderr, "recognize:", event.Name)
						RunFileRecognize([]string{event.Name})
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	fatalIfError(err)
	<-done
}
