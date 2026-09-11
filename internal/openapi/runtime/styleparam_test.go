// Copyright 2021 DeepMap, Inc.
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

import "testing"

func TestStyleParamWithOptionsByteFormat(t *testing.T) {
	got, err := StyleParamWithOptions("form", true, "payload", []byte("hello"), StyleParamOptions{
		ParamLocation: ParamLocationQuery,
		Format:        "byte",
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "payload=aGVsbG8%3D"; got != want {
		t.Fatalf("StyleParamWithOptions() = %q, want %q", got, want)
	}
}

func TestStyleParamWithOptionsAllowReserved(t *testing.T) {
	got, err := StyleParamWithOptions("form", true, "query[]", "a/b?c=d e", StyleParamOptions{
		ParamLocation: ParamLocationQuery,
		AllowReserved: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "query%5B%5D=a/b?c=d%20e"; got != want {
		t.Fatalf("StyleParamWithOptions() = %q, want %q", got, want)
	}
}
