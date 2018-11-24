package main

import (
	"encoding/json"
	"fmt"
	"github.com/kfirz/gitzup/internal/common"
	"io/ioutil"
	"log"
	"os"

	"github.com/kfirz/gitzup/internal/gcp/assets"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Applies GCP project resources.",
	Long:  "GCP project resource",
}

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a GCP project resource.",
	Run: func(cmd *cobra.Command, args []string) {

		initResponse := &common.ResourceInitResponse{
			StateAction: common.Action{
				Cmd: []string{"state"},
			},
		}

		projectSchemaBytes, err := assets.Asset("schema/project.json")
		if err != nil {
			log.Fatal("failed loading GCP project schema")
		}

		err = json.Unmarshal(projectSchemaBytes, &initResponse.ConfigSchema)
		if err != nil {
			log.Fatal("failed loading GCP project schema")
		}

		json, err := json.Marshal(&initResponse)
		if err != nil {
			log.Fatal("failed loading GCP project schema")
		}

		err = os.MkdirAll("/gitzup", 0755)
		if err != nil {
			log.Fatal("failed creating /gitzup")
		}

		err = ioutil.WriteFile("/gitzup/result.json", json, 0644)
		fmt.Println("project called")
	},
}

var projectStateCmd = &cobra.Command{
	Use:   "state",
	Short: "Discovers the state of the GCP project.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("project called")
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectStateCmd)
}
