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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type bayMoveServer struct {
	mu                sync.Mutex
	devices           map[string]uuid.UUID
	racks             map[string]uuid.UUID
	bays              map[uuid.UUID]remoteDeviceBay
	requests          []capturedRequest
	pageSize          int
	listStatus        int
	destinationStatus int
	deviceStatus      int
	patchStatus       int
	patchFailAt       int
	patchAttempts     int
	invalidJSON       bool
	invalidBay        bool
	emptyNext         bool
}

func (s *bayMoveServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.requests = append(s.requests, capturedRequest{method: r.Method, path: r.URL.Path, body: body})
	switch {
	case strings.Contains(r.URL.Path, "/dcim/device-bays/"):
		s.handleBay(w, r, body)
	case strings.Contains(r.URL.Path, "/dcim/devices/"):
		s.handleDevice(w, r)
	case strings.Contains(r.URL.Path, "/dcim/racks/"):
		s.handleRack(w, r)
	default:
		_, _ = io.WriteString(w, emptyListJSON)
	}
}

func (s *bayMoveServer) handleBay(w http.ResponseWriter, r *http.Request, body []byte) {
	switch r.Method {
	case http.MethodGet:
		s.listBays(w, r.URL.Query())
	case http.MethodPatch:
		s.patchBay(w, r, body)
	default:
		http.Error(w, "unexpected bay method", http.StatusMethodNotAllowed)
	}
}

func (s *bayMoveServer) listBays(w http.ResponseWriter, query url.Values) {
	if s.listStatus != 0 {
		http.Error(w, "list failure", s.listStatus)
		return
	}
	if s.destinationStatus != 0 && query.Has("name") {
		http.Error(w, "destination failure", s.destinationStatus)
		return
	}
	if s.invalidJSON {
		_, _ = io.WriteString(w, `{`)
		return
	}
	s.writeBayPage(w, query)
}

func (s *bayMoveServer) writeBayPage(w http.ResponseWriter, query url.Values) {
	var matches []remoteDeviceBay
	for _, bay := range s.bays {
		if matchesBayFilter(bay, query) {
			matches = append(matches, bay)
		}
	}
	slices.SortFunc(matches, func(a, b remoteDeviceBay) int { return strings.Compare(a.id.String(), b.id.String()) })
	offset, _ := strconv.Atoi(query.Get("offset"))
	end := len(matches)
	if s.pageSize > 0 {
		end = min(end, offset+s.pageSize)
	}
	results := make([]map[string]any, 0, end-offset)
	for _, bay := range matches[offset:end] {
		data := bayMoveJSON(bay)
		if s.invalidBay {
			delete(data, "id")
		}
		results = append(results, data)
	}
	var next *string
	if end < len(matches) {
		value := "?offset=" + strconv.Itoa(end)
		next = &value
	}
	if s.emptyNext {
		value := "?offset=1"
		next, results = &value, nil
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"count": len(matches), "next": next, "results": results})
}

func matchesBayFilter(bay remoteDeviceBay, query url.Values) bool {
	if query.Has("installed_device") && !slices.Contains(query["installed_device"], bay.installed.String()) {
		return false
	}
	if query.Has("device") && !slices.Contains(query["device"], bay.parent.String()) {
		return false
	}
	return !query.Has("name") || slices.Contains(query["name"], bay.name)
}

func (s *bayMoveServer) patchBay(w http.ResponseWriter, r *http.Request, body []byte) {
	s.patchAttempts++
	if s.patchStatus != 0 && (s.patchFailAt == 0 || s.patchAttempts == s.patchFailAt) {
		http.Error(w, "patch failure", s.patchStatus)
		return
	}
	id, err := uuid.Parse(path.Base(r.URL.Path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var patch struct {
		InstalledDevice *struct {
			ID uuid.UUID `json:"id"`
		} `json:"installed_device"`
	}
	if err := json.Unmarshal(body, &patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bay := s.bays[id]
	installed := uuid.Nil
	if patch.InstalledDevice != nil {
		installed = patch.InstalledDevice.ID
	}
	if err := s.validateBayInstall(bay, installed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bay.installed = installed
	s.bays[id] = bay
	_ = json.NewEncoder(w).Encode(bayMoveJSON(bay))
}

func (s *bayMoveServer) validateBayInstall(bay remoteDeviceBay, installed uuid.UUID) error {
	if installed == uuid.Nil {
		return nil
	}
	if bay.installed != uuid.Nil && bay.installed != installed {
		return fmt.Errorf("bay already occupied")
	}
	for _, existing := range s.bays {
		if existing.id != bay.id && existing.installed == installed {
			return fmt.Errorf("device already installed in another bay")
		}
	}
	return nil
}

func (s *bayMoveServer) handleDevice(w http.ResponseWriter, r *http.Request) {
	if s.deviceStatus != 0 {
		http.Error(w, "device lookup failure", s.deviceStatus)
		return
	}
	if r.URL.Query().Has("name") {
		results := []map[string]any{}
		if id, ok := s.devices[r.URL.Query().Get("name")]; ok {
			results = append(results, map[string]any{"id": id, "name": r.URL.Query().Get("name"), "display": "device"})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"count": len(results), "results": results})
		return
	}
	id, err := uuid.Parse(path.Base(r.URL.Path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPatch && s.deviceInstalled(id) {
		http.Error(w, "uninstall child before updating placement", http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "name": "device", "display": "device"})
}

func (s *bayMoveServer) deviceInstalled(id uuid.UUID) bool {
	for _, bay := range s.bays {
		if bay.installed == id {
			return true
		}
	}
	return false
}

func (s *bayMoveServer) handleRack(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	results := []map[string]any{}
	if id, ok := s.racks[name]; ok {
		results = append(results, map[string]any{"id": id, "name": name, "display": name})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"count": len(results), "results": results})
}

func (s *bayMoveServer) snapshot() ([]remoteDeviceBay, []capturedRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bays := make([]remoteDeviceBay, 0, len(s.bays))
	for _, bay := range s.bays {
		bays = append(bays, bay)
	}
	slices.SortFunc(bays, func(a, b remoteDeviceBay) int { return strings.Compare(a.id.String(), b.id.String()) })
	return bays, slices.Clone(s.requests)
}

func bayMoveJSON(bay remoteDeviceBay) map[string]any {
	data := map[string]any{
		"id": bay.id, "name": bay.name, "device": map[string]any{"id": bay.parent},
	}
	if bay.installed != uuid.Nil {
		data["installed_device"] = map[string]any{"id": bay.installed}
	}
	return data
}
