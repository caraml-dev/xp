package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/gojek/turing-experiments/management-service/server"
)

var cfgFile []string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Experiment Management server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := server.NewServer(cfgFile)
		if err != nil {
			log.Fatal(err)
		}
		server.Start()
	},
}

func init() {
	serveCmd.Flags().StringArrayVar(&cfgFile, "config", []string{},
		`Path to one or more configuration files. The flag can be set multiple times
	and the later values will take precedence.`)
	RootCmd.AddCommand(serveCmd)
}
