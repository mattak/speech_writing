package main

import "github.com/spf13/cobra"

var (
	RootCmd = &cobra.Command{
		Use:     "root",
		Short:   "speech writing tools",
		Long:    "speech writing tools",
		Version: VERSION,
	}
)

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.AddCommand(RecordSplitCmd)
	RootCmd.AddCommand(WatchRecognizeFilesCmd)
}
