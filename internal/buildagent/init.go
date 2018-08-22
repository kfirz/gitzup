package buildagent

import "log"

func (resource *Resource) Initialize() error {
	log.Printf("Initializing resource '%s'...", resource)
	return nil
}
