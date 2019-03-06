package main

import (
	"github.com/spf13/cobra"
)

var cmdTest2 = &cobra.Command{
	Use:    "test2",
	Short:  "Second Test Command",
	Run:    test2Command,
	Args:   cobra.NoArgs,
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(cmdTest2)
}

func test2Command(cmd *cobra.Command, args []string) {
	cfg := zapConfig()
	logger, _ := cfg.Build()
	defer logger.Sync()

	// httpFs := afero.NewHttpFs(webasset.Fs)
	// http.Handle("/", http.StripPrefix("/", http.FileServer(httpFs.Dir("/"))))
	// http.ListenAndServe(":8080", nil)
}
