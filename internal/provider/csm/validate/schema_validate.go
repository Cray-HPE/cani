/*
MIT License

(C) Copyright 2023 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package validate

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	"github.com/rs/zerolog/log"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func loadSchema(filename string) (schema *jsonschema.Schema, err error) {
	files, err := fs.ReadDir(schemas, "schemas")
	if err != nil {
		return
	}
	for _, file := range files {
		if filename == file.Name() {

			filePath := fmt.Sprintf("schemas/%s", file.Name())
			content, err := schemas.ReadFile(filePath)
			if err != nil {
				log.Error().Msgf("Error reading embeded schema file: %s. %s\n", filePath, err)
				return nil, err
			}
			schema, err := jsonschema.CompileString(file.Name(), string(content))
			if err != nil {
				log.Error().Msgf("Error compiling embeded schema. file: %s. %s\n", filePath, err)
				return nil, err
			}
			return schema, nil
		}
	}
	err = fmt.Errorf("failed to find schema file: %s", filename)
	return nil, err
}

func schemaValidationErrors(instancePrefix string, err *jsonschema.ValidationError) []common.ValidationResult {
	r := make([]common.ValidationResult, 0)
	if err.Message != "" {
		r = append(r,
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: instancePrefix + err.InstanceLocation,
				Description: err.Message})
	}
	for _, c := range err.Causes {
		rr := schemaValidationErrors(instancePrefix, c)
		r = append(r, rr...)
	}
	return r
}

func toValidationErrors(instancePrefix string, err error) []common.ValidationResult {
	var jsonerr *jsonschema.ValidationError
	r := make([]common.ValidationResult, 0)
	switch {
	case errors.As(err, &jsonerr):
		if len(jsonerr.Causes) == 0 {
			// todo verify that the top level error is alwasy generic when there are causes
			r = append(r,
				common.ValidationResult{
					CheckID:     common.SLSSchemaCheck,
					Result:      common.Fail,
					ComponentID: instancePrefix + jsonerr.InstanceLocation,
					Description: jsonerr.Message})
		}
		for _, c := range jsonerr.Causes {
			rr := schemaValidationErrors(instancePrefix, c)
			r = append(r, rr...)
		}
	default:
		r = append(r,
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS Networks error validating schema. %#v", err)})
	}

	return r
}

func validateSchema(schema *jsonschema.Schema, rawJson map[string]interface{}, id string, description string) []common.ValidationResult {
	results := make([]common.ValidationResult, 0)

	if err := schema.Validate(rawJson); err != nil {
		r := toValidationErrors(id, err)
		return append(results, r...)
	}

	r := common.ValidationResult{
		CheckID:     common.SLSSchemaCheck,
		Result:      common.Pass,
		ComponentID: id,
		Description: description}
	return append(results, r)
}

func validateSchemaNetworks(schema *jsonschema.Schema, networks map[string]interface{}) []common.ValidationResult {
	return validateSchema(schema, networks, "Networks", "SLS Networks is valid json")
}

// validateAgainstSchemas validates the SLS response against the schemas
func validateAgainstSchemas(slsDump RawJson) []common.ValidationResult {
	results := make([]common.ValidationResult, 0)

	networks, found := common.GetMap(slsDump, "Networks")
	if !found {
		results = append(results,
			common.ValidationResult{
				CheckID:     common.SLSSchemaCheck,
				Result:      common.Fail,
				ComponentID: "SLS Networks",
				Description: "Failed to find Networks data in SLS dump."})
		return results
	}

	networksSchema, err := loadSchema("sls_networks_schema.json")
	if err != nil {
		log.Error().Msg(err.Error())
		return results
	}

	r := validateSchemaNetworks(networksSchema, networks)
	results = append(results, r...)
	return results
}
