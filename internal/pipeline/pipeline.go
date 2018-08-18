package pipeline

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

type Pipeline struct {
	Version     int
	Name        string
	Timeout     int
	Environment string
	Strands     []*Strand
	Artifacts   []*Artifact
}

func (pipeline *Pipeline) getStrand(name string) *Strand {
	for _, s := range pipeline.Strands {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (pipeline *Pipeline) Build() error {
	for _, strand := range pipeline.Strands {
		err := strand.initialize(pipeline)
		if err != nil {
			return err
		}
	}

	var strands = sync.WaitGroup{}
	strands.Add(len(pipeline.Strands))

	for _, strand := range pipeline.Strands {
		go func(strand *Strand) {
			err := strand.build()
			if err != nil {
				fmt.Printf("Strand '%s' FAILED: %s\n", strand.Name, err.Error())
			} else {
				fmt.Printf("Strand '%s' SUCCEEDED\n", strand.Name)
			}
			strands.Done()
		}(strand)
	}

	strands.Wait()
	return nil
}

func ParsePipeline(filename string) (pipeline Pipeline, err error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(source, &pipeline)
	if err != nil {
		return
	}

	return pipeline, nil
}
