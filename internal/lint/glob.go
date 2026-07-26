package lint

import (
	"regexp"
	"strings"
)

// glob is a compiled path glob. Supported syntax:
//
//	"*"  matches any run of non-separator characters;
//	"?"  matches a single non-separator character;
//	"**" matches any run including separators;
//	"**/" matches zero or more leading path segments.
//
// Paths are always compared in slash form.
type glob struct {
	re  *regexp.Regexp
	raw string
}

func compileGlob(pattern string) glob {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					b.WriteString("(?:.*/)?")
					i += 2
					continue
				}
				b.WriteString(".*")
				i++
				continue
			}
			b.WriteString("[^/]*")
		case '?':
			b.WriteString("[^/]")
		default:
			if strings.IndexByte(`.+()|[]{}^$\`, c) >= 0 {
				b.WriteByte('\\')
			}
			b.WriteByte(c)
		}
	}
	b.WriteString("$")
	return glob{re: regexp.MustCompile(b.String()), raw: pattern}
}

func (g glob) match(path string) bool { return g.re.MatchString(path) }

// matchesAny reports whether path matches any of the compiled globs.
func matchesAny(globs []glob, path string) bool {
	for _, g := range globs {
		if g.match(path) {
			return true
		}
	}
	return false
}

// coversMatch reports whether a spec's `covers` entry designates the given
// file. A plain path matches the file itself or anything beneath it (so a
// directory like `internal/skill` covers every file under it); an entry with
// glob metacharacters is matched as a glob.
func coversMatch(entry, file string) bool {
	if strings.ContainsAny(entry, "*?") {
		return compileGlob(entry).match(file)
	}
	entry = strings.TrimSuffix(entry, "/")
	return file == entry || strings.HasPrefix(file, entry+"/")
}
