package buildagent

import "log"

func (request *BuildRequest) Apply() error {
	log.Printf("Applying build request '%#v'", request)

	for _, resource := range request.Resources {
		resource.Initialize()
	}

	for _, resource := range request.Resources {
		resource.DiscoverState()
	}

	return nil
}
