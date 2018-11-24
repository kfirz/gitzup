package assets

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-errors/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Compiled JSON schema.
type Schema struct {
	jsonLoader       *gojsonschema.JSONLoader
	underlyingSchema *gojsonschema.Schema
}

// Shortcut for loading a JSON schema, and panic-ing on errors
func LoadSchema(mainSchema interface{}, additionalSchemas ...interface{}) *Schema {
	schema, err := NewSchema(mainSchema, additionalSchemas...)
	if err != nil {
		panic(err)
	}
	return schema
}

// Creates a JSON schema JSON loader.
func newJSONLoader(source interface{}) (*gojsonschema.JSONLoader, error) {
	switch source := source.(type) {
	case string:
		schemaBytes, err := Asset(source)
		if err != nil {
			return nil, err
		}
		jsonLoader, err := newJSONLoader(schemaBytes)
		if err != nil {
			return nil, err
		}
		return jsonLoader, nil
	case *string:
		schemaBytes, err := Asset(*source)
		if err != nil {
			return nil, err
		}
		jsonLoader, err := newJSONLoader(schemaBytes)
		if err != nil {
			return nil, err
		}
		return jsonLoader, nil
	case []byte:
		jsonLoader := gojsonschema.NewBytesLoader(source)
		return &jsonLoader, nil
	case *[]byte:
		jsonLoader := gojsonschema.NewBytesLoader(*source)
		return &jsonLoader, nil
	case Schema:
		return source.jsonLoader, nil
	case *Schema:
		return (*source).jsonLoader, nil
	default:
		jsonLoader := gojsonschema.NewGoLoader(source)
		return &jsonLoader, nil
	}
}

// Parse & compile a JSON schema from the given sources. The main schema is provided as the first argument, separated
// from any additional schemas that it may reference, that are provided in the second varargs argument.
//
// Each source may be one of:
//  - *string, string: path to an embedded asset. This is used as Asset(<value>)
//  - *[]byte, []byte: bytes containing the actual JSON schema source code
func NewSchema(mainSchemaSource interface{}, additionalSchemaSources ...interface{}) (*Schema, error) {

	// create the main schema loader
	schemaLoader := gojsonschema.NewSchemaLoader()

	// add the extra schemas (these should not include the entrypoint "main" schema)
	for _, source := range additionalSchemaSources {
		jsonLoader, err := newJSONLoader(source)
		if err != nil {
			return nil, err
		}
		err = schemaLoader.AddSchemas(*jsonLoader)
		if err != nil {
			return nil, err
		}
	}

	// compile the full schema
	jsonLoader, err := newJSONLoader(mainSchemaSource)
	if err != nil {
		return nil, err
	}
	underlyingSchema, err := schemaLoader.Compile(*jsonLoader)
	if err != nil {
		return nil, err
	}
	return &Schema{jsonLoader: jsonLoader, underlyingSchema: underlyingSchema}, nil
}

// Parse the JSON from the given source bytes into the given target object.
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

// Validate that the given source complies with this schema.
func (schema *Schema) Validate(source interface{}) (err error) {
	var result *gojsonschema.Result

	switch source := source.(type) {
	case string:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewStringLoader(source))
	case *string:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewStringLoader(*source))
	case []byte:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewBytesLoader(source))
	case *[]byte:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewBytesLoader(*source))
	default:
		result, err = schema.underlyingSchema.Validate(gojsonschema.NewGoLoader(source))
	}

	if err != nil {
		return err
	} else if !result.Valid() {
		var msg = ""
		for _, e := range result.Errors() {
			msg += fmt.Sprintf("\t- %s\n", e.String())
		}
		return errors.New(fmt.Sprintf("JSON validation failed:\n%s", msg))
	}

	return nil
}
