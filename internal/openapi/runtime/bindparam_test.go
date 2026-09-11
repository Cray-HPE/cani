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

import (
	"net/url"
	"testing"
)

func TestBindStyledParameterWithOptionsByteFormat(t *testing.T) {
	var got []byte
	err := BindStyledParameterWithOptions("simple", "payload", "aGVsbG8=", &got, BindStyledParameterOptions{
		Required: true,
		Format:   "byte",
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "hello"; string(got) != want {
		t.Fatalf("BindStyledParameterWithOptions() = %q, want %q", got, want)
	}
}

func TestBindQueryParameterWithOptionsByteFormat(t *testing.T) {
	var got []byte
	err := BindQueryParameterWithOptions("form", true, true, "payload", url.Values{
		"payload": {"aGVsbG8="},
	}, &got, BindQueryParameterOptions{Format: "byte"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "hello"; string(got) != want {
		t.Fatalf("BindQueryParameterWithOptions() = %q, want %q", got, want)
	}
}

func TestBindQueryParameterWithOptionsRejectsInvalidByteFormat(t *testing.T) {
	var got []byte
	err := BindQueryParameterWithOptions("form", true, true, "payload", url.Values{
		"payload": {"not base64"},
	}, &got, BindQueryParameterOptions{Format: "byte"})
	if err == nil {
		t.Fatal("BindQueryParameterWithOptions() error = nil, want base64 error")
	}
}
