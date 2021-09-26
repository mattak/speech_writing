package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	WatchUpdateEsaCmd = &cobra.Command{
		Use:     "watch_update_esa",
		Short:   "update esa by watching file",
		Long:    "update esa by watching file",
		Example: " speech_writing watch_update_esa myteam 1234 /tmp/records/log.md",
		Run:     runCommandWatchUpdateEsa,
	}
)

func runCommandWatchUpdateEsa(cmd *cobra.Command, args []string) {
	if len(args) < 3 {
		log.Fatal("usage: <team> <post_number> <watch_file>")
	}
	token := os.Getenv("ESA_TOKEN")
	team := args[0]
	postNumber, err := strconv.ParseInt(args[1], 10, 64)
	fatalIfError(err)
	watchFile := args[2]
	RunWatchUpdateEsa(token, team, int(postNumber), watchFile)
}

func patchEsa(accessToken string, team string, postNumber int, fileName string) {
	rawData, err := ioutil.ReadFile(fileName)
	fatalIfError(err)
	patchBody := EsaPatchBody{
		BodyMd: string(rawData),
	}
	_, err = PatchDoc(accessToken, team, postNumber, patchBody)
	fatalIfError(err)
}

func RunWatchUpdateEsa(token string, team string, postNumber int, watchFile string) {
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					if strings.HasSuffix(event.Name, ".md") {
						fmt.Fprintln(os.Stderr, "update:", event.Name)
						patchEsa(token, team, postNumber, watchFile)
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

	err = watcher.Add(watchFile)
	fatalIfError(err)
	<-done
}
