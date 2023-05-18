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
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func pos(length int) int {
	if length < 0 {
		return 0
	}
	return length
}

func toSliceOfMaps(path []string, object interface{}) ([]map[string]interface{}, bool) {
	slice, found := toSlice(path, object)
	if found {
		result := make([]map[string]interface{}, len(slice))
		for i, s := range slice {
			result[i], _ = toMap([]string{}, s)
		}
		return result, true
	}
	return make([]map[string]interface{}, 0), false
}

func toSlice(path []string, object interface{}) ([]interface{}, bool) {
	mapPath := path[:pos(len(path)-1)]
	objectMap, found := toMap(mapPath, object)

	if found {
		key := path[len(path)-1]
		if value, found := objectMap[key]; found {
			if slice, ok := value.([]interface{}); ok {
				return slice, true
			}
		}
	}
	return make([]interface{}, 0), false
}

func toMap(path []string, object interface{}) (map[string]interface{}, bool) {
	if len(path) == 0 {
		if m, ok := object.(map[string]interface{}); ok {
			return m, true
		}
	}

	for _, key := range path {
		if m, ok := object.(map[string]interface{}); ok {
			if value, found := m[key]; found {
				object = value
			} else {
				return make(map[string]interface{}), false
			}
		}
	}
	return object.(map[string]interface{}), true
}

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
				fmt.Printf("Error reading embeded schema file: %s. %s\n", filePath, err)
				return nil, err
			}
			schema, err := jsonschema.CompileString(file.Name(), string(content))
			if err != nil {
				fmt.Printf("Error compiling embeded schema. file: %s. %s\n", filePath, err)
				return nil, err
			}
			return schema, nil
		}
	}
	err = fmt.Errorf("failed to find schema file: %s", filename)
	return nil, err
}

func validateNetworksSchema(schema *jsonschema.Schema, networks map[string]interface{}) []ValidationResult {
	results := make([]ValidationResult, 0)

	if err := schema.Validate(networks); err != nil {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS Networks error validating schema. %s", err)})
		return results
	}
	results = append(results,
		ValidationResult{
			CheckID:     SLSSchemaCheck,
			Result:      Pass,
			ComponentID: "SLS Networks",
			Description: "SLS Networks valid json"})

	return results
}

// func validateReservationsSchema(schema *jsonschema.Schema, reservation map[string]interface{}) []ValidationResult {
func validateReservationsSchema(schema *jsonschema.Schema, reservation []map[string]interface{}, id *ID) []ValidationResult {
	results := make([]ValidationResult, 0)
	if err := schema.Validate(reservation); err != nil {
		results = append(results,
			ValidationResult{
				CheckID: SLSSchemaCheck,
				Result:  Fail,
				// ComponentID: "SLS Networks",
				ComponentID: "SLS Networks: " + id.str(),
				Description: fmt.Sprintf("SLS Networks error validating schema. %s", err)})
		return results
	}
	results = append(results,
		ValidationResult{
			CheckID: SLSSchemaCheck,
			Result:  Pass,
			// ComponentID: "SLS Networks",
			ComponentID: "SLS Networks: " + id.str(),
			Description: "SLS Networks valid json"})
	return results
}

func validateSubnetsSchema(schema *jsonschema.Schema, subnet map[string]interface{}) []ValidationResult {
	results := make([]ValidationResult, 0)
	if err := schema.Validate(subnet); err != nil {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS Networks error validating schema. %s", err)})
		return results
	}
	results = append(results,
		ValidationResult{
			CheckID:     SLSSchemaCheck,
			Result:      Pass,
			ComponentID: "SLS Networks",
			Description: "SLS Networks valid json"})
	return results
}

func validateAgainstSchemas(response *http.Response) []ValidationResult {
	results := make([]ValidationResult, 0)

	responseBytes, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS failed to get raw json jumpstate. %s", err)})
		return results
	}

	var slsDump interface{}
	if err := json.Unmarshal(responseBytes, &slsDump); err != nil {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS Networks",
				Description: fmt.Sprintf("SLS Networks error unmarshaling json. %s", err)})
		return results
	}

	networks, found := toMap([]string{"Networks"}, slsDump)
	if !found {
		results = append(results,
			ValidationResult{
				CheckID:     SLSSchemaCheck,
				Result:      Fail,
				ComponentID: "SLS Networks",
				Description: "Failed to find Networks data in SLS dump."})
		return results
	}

	networksSchema, err := loadSchema("sls_networks_schema.json")
	if err != nil {
		fmt.Println(err)
		return results
	}

	reservationsSchema, err := loadSchema("sls_reservations_schema.json")
	if err != nil {
		fmt.Println(err)
		return results
	}

	subnetsSchema, err := loadSchema("sls_subnets_schema.json")
	if err != nil {
		fmt.Println(err)
		return results
	}

	r := validateNetworksSchema(networksSchema, networks)
	results = append(results, r...)
	for name, networkRaw := range networks {
		networkId := NewID("Network", name)
		fmt.Println(networkId.strYaml())
		subnets, found := toSliceOfMaps([]string{"ExtraProperties", "Subnets"}, networkRaw)
		if found {
			for _, subnet := range subnets {
				cidr := subnet["CIDR"]
				subnetId := networkId.append(Pair{"SubnetCIDR", cidr.(string)})
				fmt.Println(subnetId.strYaml())

				// r := validateReservationsSchema(reservationsSchema, subnet)
				r := validateSubnetsSchema(subnetsSchema, subnet)
				results = append(results, r...)
				reservations, found := toSliceOfMaps([]string{"IPReservations"}, subnet)
				if found {
					r := validateReservationsSchema(reservationsSchema, reservations, subnetId)
					results = append(results, r...)
					// 	for _, reservation := range reservations {
					// 		r := validateSubnetsSchema(subnetsSchema, reservation)
					// 		results = append(results, r...)
					// 	}
				}
			}

		}
	}

	return results
}
