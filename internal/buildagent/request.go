package buildagent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kfirz/gitzup/internal/assets"
	"github.com/xeipuuv/gojsonschema"
)

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
	// TODO: only build JSON schema one (instead of for every request)

	buildRequestJsonSchema, err := assets.Asset("build-request.schema.json")
	if err != nil {
		return nil, err
	}

	resourceJsonSchema, err := assets.Asset("resource.schema.json")
	if err != nil {
		return nil, err
	}

	// create the main schema loader
	schemaLoader := gojsonschema.NewSchemaLoader()

	// add schema fragments
	err = schemaLoader.AddSchemas(gojsonschema.NewBytesLoader(resourceJsonSchema))
	if err != nil {
		return nil, err
	}

	// compile the full schema
	schema, err := schemaLoader.Compile(gojsonschema.NewBytesLoader(buildRequestJsonSchema))
	if err != nil {
		return nil, err
	}

	// validate the given document
	validationResult, err := schema.Validate(gojsonschema.NewBytesLoader(b))
	if err != nil {
		return nil, err
	} else if !validationResult.Valid() {
		var msg = ""
		for _, e := range validationResult.Errors() {
			msg += fmt.Sprintf("\t- %s\n", e.String())
		}
		return nil, errors.New(fmt.Sprintf("Invalid request:\n%s", msg))
	}

	// JSON is valid; translate to BuildRequest instance
	var req BuildRequest
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	err = decoder.Decode(&req)
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
