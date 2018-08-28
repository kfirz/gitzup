package buildagent

import (
	"fmt"
	"github.com/kfirz/gitzup/internal/util"
	"log"
)

var buildRequestSchema *util.Schema

func init() {
	schema, err := util.NewSchema("build-request.schema.json", "resource.schema.json")
	if err != nil {
		panic(err)
	}
	buildRequestSchema = schema
}

type Resource struct {
	Request BuildRequest           `json:"-"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
}

type BuildRequest struct {
	Id        string
	Resources map[string]Resource `json:"resources"`
}

func (request *BuildRequest) WorkspacePath() string {
	// TODO: make the build-request's workspace path absolute
	return fmt.Sprintf("./%s", request.Id)
}

func NewBuildRequest(id string, b []byte) (*BuildRequest, error) {
	var req BuildRequest

	// validate & parse the build request
	err := buildRequestSchema.ParseAndValidate(&req, b)
	if err != nil {
		return nil, err
	}

	// post-parsing updates
	req.Id = id

	for name, resource := range req.Resources {
		resource.Name = name
		resource.Request = req
	}
	return &req, nil
}

func (request *BuildRequest) Apply() error {
	log.Printf("Applying build request '%#v'", request)

	for _, resource := range request.Resources {
		err := resource.Initialize()
		if err != nil {
			return err
		}
	}

	for _, resource := range request.Resources {
		err := resource.DiscoverState()
		if err != nil {
			return err
		}
	}

	return nil
}
