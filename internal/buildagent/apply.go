package buildagent

import "log"

func (resource *Resource) Apply() error {
	log.Printf("Applying resource '%s'...", resource)
	return nil
}
