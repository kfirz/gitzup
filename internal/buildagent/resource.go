package buildagent

import "fmt"

type Resource struct {
	Request BuildRequest           `json:"-"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
}

func (resource *Resource) Workspace() string {
	return fmt.Sprintf("%s/resources/%s", resource.Request.WorkspacePath(), resource.Name)
}
