// Package spec parses the YAML frontmatter of spec markdown files.
//
// A spec is a plain markdown file whose frontmatter declares a stable id, a
// human title, and (optionally) the source paths it governs. The prose body is
// for humans and agents; specguard never interprets it. Only a constrained
// subset of YAML is supported — scalar `key: value` pairs and one-level block
// sequences — so the tool stays dependency-free.
package spec

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Spec is one parsed spec file.
type Spec struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Status string   `json:"status,omitempty"`
	Covers []string `json:"covers,omitempty"`
	Path   string   `json:"path"`
}

// idPattern constrains ids to the characters that also survive being embedded
// in a `spec:<id>` tag inside Go comments and Playwright tags, so a reference
// can always be matched back to a definition.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// ParseFile reads and parses a single spec markdown file.
func ParseFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s, err := Parse(data)
	if err != nil {
		return nil, err
	}
	s.Path = path
	return s, nil
}

// Parse parses the frontmatter of a spec file's contents.
func Parse(data []byte) (*Spec, error) {
	fm, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}
	fields, err := ParseFields(fm)
	if err != nil {
		return nil, err
	}
	s := &Spec{Covers: fields["covers"]}
	if v := fields["id"]; len(v) > 0 {
		s.ID = v[0]
	}
	if v := fields["title"]; len(v) > 0 {
		s.Title = v[0]
	}
	if v := fields["status"]; len(v) > 0 {
		s.Status = v[0]
	}
	if s.ID == "" {
		return nil, fmt.Errorf("frontmatter is missing required field 'id'")
	}
	if !idPattern.MatchString(s.ID) {
		return nil, fmt.Errorf("id %q must match %s", s.ID, idPattern)
	}
	if s.Title == "" {
		return nil, fmt.Errorf("frontmatter is missing required field 'title'")
	}
	return s, nil
}

// splitFrontmatter returns the text between the opening and closing `---`
// fences at the top of the document.
func splitFrontmatter(data []byte) (string, error) {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return "", fmt.Errorf("file must start with a '---' frontmatter fence")
	}
	rest := s[len("---\n"):]
	// The closing fence is a line that is exactly "---".
	if i := strings.Index(rest, "\n---"); i >= 0 {
		// Ensure the match is a whole line ("---" followed by newline or EOF).
		tail := rest[i+len("\n---"):]
		if tail == "" || strings.HasPrefix(tail, "\n") {
			return rest[:i], nil
		}
	}
	return "", fmt.Errorf("unterminated frontmatter (missing closing '---')")
}

// ParseFields parses the constrained YAML subset used by both spec frontmatter
// and the config file: scalar `key: value` pairs and one-level block sequences.
// Every value is returned as a slice so scalars and lists share one shape.
func ParseFields(text string) (map[string][]string, error) {
	fields := map[string][]string{}
	curKey := ""
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Block-sequence item: "- value", only valid under a pending key.
		if trimmed == "-" || strings.HasPrefix(trimmed, "- ") {
			if curKey == "" {
				return nil, fmt.Errorf("list item %q has no parent key", trimmed)
			}
			item := unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "-")))
			if item != "" {
				fields[curKey] = append(fields[curKey], item)
			}
			continue
		}
		// "key: value" or "key:" (value on following list lines).
		idx := strings.Index(line, ":")
		if idx < 0 {
			return nil, fmt.Errorf("invalid line (expected 'key: value'): %q", line)
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			return nil, fmt.Errorf("empty key in line: %q", line)
		}
		val := strings.TrimSpace(line[idx+1:])
		if val == "" {
			// A block sequence for this key may follow.
			curKey = key
			if _, ok := fields[key]; !ok {
				fields[key] = nil
			}
			continue
		}
		fields[key] = []string{unquote(val)}
		curKey = "" // a scalar ends any list context
	}
	return fields, nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
