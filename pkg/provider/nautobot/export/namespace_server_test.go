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
package export

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// namespaceServer imitates Nautobot's IPAM endpoints closely enough for
// namespace tests: only Global exists until a namespace is POSTed, a list
// filtered by an unknown namespace answers 400 as Nautobot 3.2 does, creates
// get a fresh ID, and every request is recorded.
type namespaceServer struct {
	ids      map[string]uuid.UUID
	known    map[string]bool
	requests []string
	posts    map[string][][]byte
}

func newNamespaceServer() *namespaceServer {
	return &namespaceServer{
		ids:   map[string]uuid.UUID{},
		known: map[string]bool{"Global": true},
		posts: map[string][][]byte{},
	}
}

func (s *namespaceServer) id(name string) uuid.UUID {
	if id, ok := s.ids[name]; ok {
		return id
	}
	s.ids[name] = uuid.New()
	return s.ids[name]
}

func (s *namespaceServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.requests = append(s.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	query := r.URL.Query()
	switch {
	case strings.Contains(r.URL.Path, "ipam/namespaces"):
		s.serveNamespaces(w, r)
	case r.Method == http.MethodPost:
		body, _ := io.ReadAll(r.Body)
		s.posts[r.URL.Path] = append(s.posts[r.URL.Path], body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"`+uuid.New().String()+`"}`)
	case query.Has("namespace") && !s.known[query.Get("namespace")]:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"namespace":["Select a valid choice."]}`)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

// serveNamespaces finds a known namespace by name or creates one from a POST.
func (s *namespaceServer) serveNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var sent struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&sent)
		s.known[sent.Name] = true
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, s.namespaceJSON(sent.Name))
		return
	}
	name := r.URL.Query().Get("name")
	if !s.known[name] {
		_, _ = io.WriteString(w, emptyListJSON)
		return
	}
	_, _ = io.WriteString(w, `{"count":1,"results":[`+s.namespaceJSON(name)+`]}`)
}

func (s *namespaceServer) namespaceJSON(name string) string {
	return `{"id":"` + s.id(name).String() + `","name":"` + name + `","display":"` + name + `"}`
}

func (s *namespaceServer) requestsMatching(substr string) []string {
	var matches []string
	for _, request := range s.requests {
		if strings.Contains(request, substr) {
			matches = append(matches, request)
		}
	}
	return matches
}

// writes returns every request that was not a GET.
func (s *namespaceServer) writes() []string {
	var writes []string
	for _, request := range s.requests {
		if !strings.HasPrefix(request, http.MethodGet+" ") {
			writes = append(writes, request)
		}
	}
	return writes
}
