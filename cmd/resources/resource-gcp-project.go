package main

import (
	"github.com/kfirz/gitzup/internal/util"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var protocolInitRequestSchema *util.Schema

func init() {
	schema, err := util.NewSchema("protocol/init.request.schema.json")
	if err != nil {
		panic(err)
	}
	protocolInitRequestSchema = schema
}

type InitRequestResourceInfo struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type InitRequest struct {
	Resource InitRequestResourceInfo `json:"resource"`
}

func NewInitRequest(input io.Reader) (*InitRequest, error) {
	inputBytes, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}

	var req InitRequest
	err = protocolInitRequestSchema.ParseAndValidate(&req, inputBytes)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func PrintUsageAndExit() {
	log.Fatalf("usage: %s <init/discover>\n", os.Args[0])
}

func main() {
	// ignore first item which is the executable name itself
	args := os.Args[1:]

	// fail if no arguments given
	if len(args) == 0 {
		PrintUsageAndExit()
	}

	// command is the first item
	switch args[0] {
	case "init":
		args := args[1:]
		if len(args) > 0 {
			PrintUsageAndExit()
		}

		request, err := NewInitRequest(os.Stdin)
		if err != nil {
			panic(err)
			log.Fatalln(err.Error())
		}
		log.Printf("%#v\n", request)

	case "discover":
		args := args[1:]
		if len(args) > 0 {
			PrintUsageAndExit()
		}

		// TODO: implement "discover" command for GCP project resource
		log.Fatalln("Discovery not implemented yet.")

	case "":
	default:
		PrintUsageAndExit()
	}
}
