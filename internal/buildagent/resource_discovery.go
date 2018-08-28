package buildagent

import "log"

func (resource *Resource) DiscoverState() error {
	log.Printf("Discovering state for resource '%s'...\n", resource.Name)
	return nil
}
