package main

import (
	"bufio"
	"context"
	"github.com/kfirz/gitzup/internal/agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Process a build request.",
	Long:  `This command will build the provided build request.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(args)
		if len(args) < 1 {
			log.Fatal("build ID is required")
		}
		if len(args) < 2 {
			log.Fatal("build file is required (use '-' for stdin)")
		}

		id := args[0]
		pipelineFile := args[1]

		var bytes []byte
		var err error
		if pipelineFile == "-" {
			stdinReader := bufio.NewReader(os.Stdin)
			bytes, err = ioutil.ReadAll(stdinReader)
		} else {
			bytes, err = ioutil.ReadFile(pipelineFile)
		}
		if err != nil {
			log.WithError(err).Fatal("failed reading pipeline")
		}

		request, err := agent.New(id, workspacePath, bytes)
		if err != nil {
			log.WithError(err).Fatal("failed creating build request")
		}

		// TODO: support timeout by using "context.WithTimeout(..)" as the context to "request.Apply(ctx)" method
		err = request.Apply(context.WithValue(context.Background(), "request", request.Id()))
		if err != nil {
			log.WithError(err).Fatal("failed applying build request")
		}

		// TODO: receive apply result and print it as text/json
	},
}

// Initializes the main package with global flags
func init() {
	rootCmd.AddCommand(buildCmd)
}
