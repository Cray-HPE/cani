/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package imprt

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestFetchVLANLocationAssignmentsPagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/ipam/vlan-location-assignments/",
		[]map[string]interface{}{{
			"location": map[string]string{"id": "11111111-1111-1111-1111-111111111111"},
			"vlan":     map[string]string{"id": "22222222-2222-2222-2222-222222222222"},
		}},
		[]map[string]interface{}{{
			"location": map[string]string{"id": "33333333-3333-3333-3333-333333333333"},
			"vlan":     map[string]string{"id": "44444444-4444-4444-4444-444444444444"},
		}}, &calls)
	defer srv.Close()

	got, err := FetchVLANLocationAssignments(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchVLANLocationAssignments: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 VLAN location assignments, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}

func TestFetchPrefixLocationAssignmentsPagination(t *testing.T) {
	var calls int32
	srv := twoPageServer(t, "/ipam/prefix-location-assignments/",
		[]map[string]interface{}{{
			"location": map[string]string{"id": "11111111-1111-1111-1111-111111111111"},
			"prefix":   map[string]string{"id": "22222222-2222-2222-2222-222222222222"},
		}},
		[]map[string]interface{}{{
			"location": map[string]string{"id": "33333333-3333-3333-3333-333333333333"},
			"prefix":   map[string]string{"id": "44444444-4444-4444-4444-444444444444"},
		}}, &calls)
	defer srv.Close()

	got, err := FetchPrefixLocationAssignments(context.Background(), newTestClient(t, srv.URL))
	if err != nil {
		t.Fatalf("FetchPrefixLocationAssignments: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 prefix location assignments, got %d", len(got))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", calls)
	}
}
