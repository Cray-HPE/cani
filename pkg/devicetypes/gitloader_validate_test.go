package devicetypes

import "testing"

func TestValidateRepoURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://github.com/netbox-community/devicetype-library.git", false},
		{"http", "http://internal-mirror.example/repo.git", false},
		{"ssh", "ssh://git@github.com/org/repo.git", false},
		{"git scheme", "git://github.com/org/repo.git", false},
		{"scp shorthand", "git@github.com:org/repo.git", false},
		{"empty", "", true},
		{"leading dash", "--upload-pack=touch /tmp/pwned", true},
		{"ext transport", "ext::sh -c 'touch /tmp/pwned'", true},
		{"file transport", "file:///etc/passwd", true},
		{"unknown scheme", "ftp://example.com/repo.git", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRepoURL(tc.url)
			if tc.wantErr && err == nil {
				t.Errorf("validateRepoURL(%q) = nil, want error", tc.url)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateRepoURL(%q) = %v, want nil", tc.url, err)
			}
		})
	}
}

func TestIsSCPSyntax(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"git@github.com:org/repo.git", true},
		{"user@host:path", true},
		{"https://github.com/org/repo.git", false},
		{"ext::sh -c 'x'", false},
		{"/local/path", false},
		{"@leading", false},
	}

	for _, tc := range cases {
		if got := isSCPSyntax(tc.url); got != tc.want {
			t.Errorf("isSCPSyntax(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}
