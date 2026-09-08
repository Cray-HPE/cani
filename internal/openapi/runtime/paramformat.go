// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
)

func isByteSlice(valueType reflect.Type) bool {
	return valueType.Kind() == reflect.Slice && valueType.Elem().Kind() == reflect.Uint8
}

func base64Decode(value string) ([]byte, error) {
	if value == "" {
		return []byte{}, nil
	}

	if strings.ContainsRune(value, '=') {
		if strings.ContainsAny(value, "-_") {
			return base64DecodeWith(base64.URLEncoding, value)
		}
		return base64DecodeWith(base64.StdEncoding, value)
	}
	if strings.ContainsAny(value, "-_") {
		return base64DecodeWith(base64.RawURLEncoding, value)
	}
	return base64DecodeWith(base64.RawStdEncoding, value)
}

func base64DecodeWith(encoding *base64.Encoding, value string) ([]byte, error) {
	decoded, err := encoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("failed to base64-decode string %q: %w", value, err)
	}
	return decoded, nil
}
