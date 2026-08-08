package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/nuzirwan/go-boilerplate/internal/config"
)

var (
	rootCmd = &cobra.Command{
		Use:   "myservice",
		Short: "Product catalog microservice",
	}
	cfg *config.Config
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
