package import_

import (
	"bufio"
	"bytes"
	"strings"
)

// iniKV is a single key=value pair, preserved in source order.
type iniKV struct {
	key string
	val string
}

// iniSection is a parsed [section] and its key/value pairs. Duplicate keys are
// retained because the cm.config format repeats keys (name=, hostname1=,
// template=) once per entry — the "shadow value" behaviour the parser relies on.
type iniSection struct {
	sectionName string
	kvs         []iniKV
}

// name returns the section name without the surrounding brackets.
func (s *iniSection) name() string { return s.sectionName }

// hasKey reports whether the section holds at least one pair with the given key.
func (s *iniSection) hasKey(key string) bool {
	for _, kv := range s.kvs {
		if kv.key == key {
			return true
		}
	}
	return false
}

// valuesFor returns every value stored under key, in source order. This mirrors
// the shadow-key lookup the cm.config format depends on.
func (s *iniSection) valuesFor(key string) []string {
	var out []string
	for _, kv := range s.kvs {
		if kv.key == key {
			out = append(out, kv.val)
		}
	}
	return out
}

// iniFile is a minimally-parsed INI document: an ordered list of sections.
type iniFile struct {
	sections []*iniSection
}

// parseINI parses INI-style data into ordered sections. It is intentionally
// minimal and supports only what the HPCM cm.config format uses: [section]
// headers, whole-line "#" and ";" comments, and key=value lines split on the
// first "=". Duplicate keys within a section are retained (shadow values), and
// surrounding whitespace is trimmed from both key and value.
func parseINI(data []byte) *iniFile {
	f := &iniFile{}
	var cur *iniSection

	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Allow long lines: a single [discover] entry can be a few KB and clusters
	// may be large. Bump the max token size well past bufio's 64KB default so
	// entries are never silently dropped.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		if line[0] == '[' && line[len(line)-1] == ']' {
			cur = &iniSection{sectionName: strings.TrimSpace(line[1 : len(line)-1])}
			f.sections = append(f.sections, cur)
			continue
		}
		if cur == nil {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		cur.kvs = append(cur.kvs, iniKV{
			key: strings.TrimSpace(line[:idx]),
			val: strings.TrimSpace(line[idx+1:]),
		})
	}
	return f
}
