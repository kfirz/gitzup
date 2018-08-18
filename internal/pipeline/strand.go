package pipeline

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const (
	StrandWaiting = iota
	StrandScheduled
	StrandRunning
	StrandFailure
	StrandSkipped
	StrandSuccess
)

type Strand struct {
	Name                string
	Directory           string
	After               []string
	Timeout             int
	Steps               []Step
	status              int
	pipeline            *Pipeline
	strandEventsChannel chan *Strand // 'll listen here to receive strand events from required strands
	listeners           []chan *Strand
	requirements        map[string]int // maps required strand name to that strand's status (eg. "otherStrand -> SUCCESS")
}

func (strand *Strand) notify(channel chan *Strand) {
	strand.listeners = append(strand.listeners, channel)
}

func (strand *Strand) initialize(pipeline *Pipeline) error {
	strand.pipeline = pipeline
	strand.status = StrandWaiting

	// initialize our required strands map
	strand.requirements = make(map[string]int)
	for _, strandName := range strand.After {
		strand.requirements[strandName] = StrandWaiting
	}

	// create a channel that will be passed to our required strands so they can notify us when they are done
	strand.strandEventsChannel = make(chan *Strand)
	for _, requiredStrandName := range strand.After {
		requiredStrand := strand.pipeline.getStrand(requiredStrandName)
		if requiredStrand != nil {
			requiredStrand.notify(strand.strandEventsChannel)
		} else {
			return errors.New(fmt.Sprintf("could not find required strand '%s'", requiredStrandName))
		}
	}

	return nil
}

func (strand *Strand) build() error {
	// on completion of this method, notify our listeners
	defer func() {
		for _, i := range strand.listeners {
			i <- strand
		}
	}()

	strand.status = StrandScheduled
	for {

		// wait for a requirement to finish
		if len(strand.requirements) > 0 {
			var requiredStrand = <-strand.strandEventsChannel
			strand.requirements[requiredStrand.Name] = requiredStrand.status
		}

		// if any required strand has failed or skipped, this strand can be considered failed or skipped as well
		var requirementsMet = true
		for name, status := range strand.requirements {
			switch status {
			case StrandWaiting, StrandScheduled, StrandRunning:
				requirementsMet = false

			case StrandFailure:
				strand.status = StrandSkipped
				return errors.New(fmt.Sprintf("skipped due to failed required strand '%s'", name))

			case StrandSkipped:
				strand.status = StrandSkipped
				return errors.New(fmt.Sprintf("skipped due to skipped required strand '%s'", name))

			case StrandSuccess:
				// do nothing - all good

			default:
				strand.status = StrandFailure
				return errors.New(fmt.Sprintf("failed due to unsupported status '%d' received from '%s'", status, name))
			}
		}

		// if all requirements are met, we can build & return
		if requirementsMet {
			strand.status = StrandRunning
			fmt.Printf("Strand '%s' is building...\n", strand.Name)
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
			randFail := rand.Intn(100)
			if randFail > 50 {
				strand.status = StrandFailure
				return errors.New(fmt.Sprintf("randomly failed (%d)", randFail))
			} else {
				strand.status = StrandSuccess
				return nil
			}
		}
	}
}
