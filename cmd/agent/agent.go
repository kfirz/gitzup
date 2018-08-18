package main

import (
	"flag"
	"fmt"
	"github.com/kfirz/gitzup/internal/pipeline"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("please specify at least one pipeline file.")
		os.Exit(1)
	}

	for i := 0; i < flag.NArg(); i++ {
		var pipelinePath = flag.Arg(i)
		p, err := pipeline.ParsePipeline(pipelinePath)
		if err != nil {
			fmt.Printf("Pipeline '%s' could not be parsed: %s\n", pipelinePath, err.Error())
			continue
		}

		err = p.Build()
		if err != nil {
			fmt.Printf("Pipeline '%s' could not be built: %s\n", pipelinePath, err.Error())
			continue
		}
	}
}
