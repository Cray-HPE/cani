/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package devicetypes

import (
	"strings"
	"testing"
)

func TestValidateUserStatusAcceptsValidStatuses(t *testing.T) {
	for _, s := range AllUserStatuses {
		got, err := ValidateUserStatus(string(s))
		if err != nil {
			t.Errorf("ValidateUserStatus(%q) returned error: %v", s, err)
		}
		if got != s {
			t.Errorf("ValidateUserStatus(%q) = %q, want %q", s, got, s)
		}
	}
}

func TestValidateUserStatusIsCaseInsensitive(t *testing.T) {
	cases := []struct {
		input string
		want  NautobotStatus
	}{
		{"planned", StatusPlanned},
		{"PLANNED", StatusPlanned},
		{"Planned", StatusPlanned},
		{"offline", StatusOffline},
		{"OFFLINE", StatusOffline},
		{"end-of-life", StatusEndOfLife},
		{"END-OF-LIFE", StatusEndOfLife},
		{"extended support", StatusExtendedSupport},
	}
	for _, tc := range cases {
		got, err := ValidateUserStatus(tc.input)
		if err != nil {
			t.Errorf("ValidateUserStatus(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ValidateUserStatus(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestValidateUserStatusRejectsStaged(t *testing.T) {
	for _, s := range []string{"Staged", "staged", "STAGED"} {
		_, err := ValidateUserStatus(s)
		if err == nil {
			t.Errorf("ValidateUserStatus(%q) should have returned error", s)
		}
	}
}

func TestValidateUserStatusAcceptsActive(t *testing.T) {
	for _, s := range []string{"Active", "active", "ACTIVE"} {
		got, err := ValidateUserStatus(s)
		if err != nil {
			t.Errorf("ValidateUserStatus(%q) returned error: %v", s, err)
			continue
		}
		if got != StatusActive {
			t.Errorf("ValidateUserStatus(%q) = %q, want %q", s, got, StatusActive)
		}
	}
}

func TestValidateUserStatusRejectsInvalidStrings(t *testing.T) {
	for _, s := range []string{"bogus", "", "running", "stopped"} {
		_, err := ValidateUserStatus(s)
		if err == nil {
			t.Errorf("ValidateUserStatus(%q) should have returned error", s)
		}
	}
}

func TestIsValidStatusIncludesActiveAndStaged(t *testing.T) {
	if !IsValidStatus("Active") {
		t.Error("IsValidStatus(Active) = false, want true")
	}
	if !IsValidStatus("Staged") {
		t.Error("IsValidStatus(Staged) = false, want true")
	}
	if !IsValidStatus("active") {
		t.Error("IsValidStatus(active) = false, want true")
	}
}

func TestIsValidStatusRejectsInvalid(t *testing.T) {
	if IsValidStatus("bogus") {
		t.Error("IsValidStatus(bogus) = true, want false")
	}
}

func TestNormalizeStatusReturnsCanonicalForm(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"active", "Active"},
		{"STAGED", "Staged"},
		{"planned", "Planned"},
		{"end-of-life", "End-of-Life"},
		{"unknown", "unknown"},
	}
	for _, tc := range cases {
		got := NormalizeStatus(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestUserStatusNamesContainsAllUserStatuses(t *testing.T) {
	names := UserStatusNames()
	for _, s := range AllUserStatuses {
		if !strings.Contains(names, string(s)) {
			t.Errorf("UserStatusNames() missing %q", s)
		}
	}
	if !strings.Contains(names, "Active") {
		t.Error("UserStatusNames() should contain Active")
	}
	if strings.Contains(names, "Staged") {
		t.Error("UserStatusNames() should not contain Staged")
	}
}

func TestAllUserStatusesExcludesStaged(t *testing.T) {
	for _, s := range AllUserStatuses {
		if s == StatusStaged {
			t.Error("AllUserStatuses should not contain Staged")
		}
	}
}

func TestAllStatusesCount(t *testing.T) {
	if len(AllStatuses) != 22 {
		t.Errorf("AllStatuses length = %d, want 22", len(AllStatuses))
	}
}

func TestAllUserStatusesCount(t *testing.T) {
	if len(AllUserStatuses) != 21 {
		t.Errorf("AllUserStatuses length = %d, want 21", len(AllUserStatuses))
	}
}
