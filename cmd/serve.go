package cmd

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a server",
}

func init() {
	serveCmd.AddCommand(serveHTTPCmd)
}
