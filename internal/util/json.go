package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kfirz/gitzup/internal/assets"
	"github.com/xeipuuv/gojsonschema"
)

type Schema struct {
	underlyingSchema *gojsonschema.Schema
}

func (schema *Schema) ParseAndValidate(target interface{}, inputBytes []byte) (err error) {
	result, err := schema.underlyingSchema.Validate(gojsonschema.NewBytesLoader(inputBytes))
	if err != nil {
		return err
	}

	if !result.Valid() {
		var msg = ""
		for _, e := range result.Errors() {
			msg += fmt.Sprintf("\t- %s\n", e.String())
		}
		return errors.New(fmt.Sprintf("JSON validation failed:\n%s", msg))
	}

	// JSON is valid; translate to BuildRequest instance
	decoder := json.NewDecoder(bytes.NewReader(inputBytes))
	decoder.UseNumber()
	err = decoder.Decode(target)
	if err != nil {
		return err
	}

	return nil
}

func NewSchema(mainSchema string, extraSchemas ...string) (*Schema, error) {

	// create the main schema loader
	schemaLoader := gojsonschema.NewSchemaLoader()

	// add the extra schemas (these should not include the entrypoint "main" schema)
	for _, asset := range extraSchemas {
		schemaBytes, err := assets.Asset(asset)
		if err != nil {
			return nil, err
		}

		err = schemaLoader.AddSchemas(gojsonschema.NewBytesLoader(schemaBytes))
		if err != nil {
			return nil, err
		}
	}

	// compile the full schema
	mainSchemaBytes, err := assets.Asset(mainSchema)
	if err != nil {
		return nil, err
	}

	underlyingSchema, err := schemaLoader.Compile(gojsonschema.NewBytesLoader(mainSchemaBytes))
	if err != nil {
		return nil, err
	}

	return &Schema{underlyingSchema: underlyingSchema}, nil
}
