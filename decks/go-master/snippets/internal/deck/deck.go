// Package deck reads code out of the Deck's segments and out of the snippet
// packages, so the tests in the module root can hold the two together.
//
// A snippet file marks the code a slide shows with a pair of comments:
//
//	// slide: 09-signals.md "Go Signals"
//	...the code exactly as the slide shows it...
//	// end slide
//
// The quoted anchor is any text in the segment before the code block, usually
// the walkthrough's `heading:`. The region must equal the first `go` block
// after it, once the region's common indentation is removed.
//
// A slide that shows several files in one block (an Activity, then the
// Workflow that calls it) is split into parts instead, each in the package
// that compiles it:
//
//	// part: 04-design-patterns.md "Fan out" 1
//	...
//	// end part
//
// The block must equal the parts in order, joined by one blank line.
package deck

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Root is the snippets module; Segments is the Deck's segments folder.
var Root, Segments = func() (string, string) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return root, filepath.Join(root, "..", "segments")
}()

var (
	slideMarker = regexp.MustCompile(`^([ \t]*)// slide: (\S+) "([^"]+)"$`)
	partMarker  = regexp.MustCompile(`^([ \t]*)// part: (\S+) "([^"]+)" (\d+)$`)
	// Fence matches a Go code block. Highlight specs and other options follow
	// the language on the same line.
	Fence = regexp.MustCompile("(?ms)^```go\\b[^\\n]*\\n(.*?)^```$")
)

// Region is one marked slide region in a snippet file.
type Region struct {
	File    string // relative to Root
	Line    int
	Segment string
	Anchor  string
	Part    int // 0 for a whole-slide region
	Code    string
}

// SegmentFiles lists the segment file names, sorted.
func SegmentFiles() ([]string, error) {
	entries, err := os.ReadDir(Segments)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// GoFiles lists every .go file in the module, relative to Root.
func GoFiles() ([]string, error) {
	var out []string
	err := filepath.WalkDir(Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != Root {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			rel, _ := filepath.Rel(Root, path)
			out = append(out, rel)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func dedent(lines []string) string {
	cut := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " \t"))
		if cut < 0 || n < cut {
			cut = n
		}
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		if cut > 0 && len(l) >= cut {
			out[i] = l[cut:]
		} else if cut > 0 {
			out[i] = "" // a blank line shorter than the indent
		} else {
			out[i] = l
		}
	}
	return strings.Join(out, "\n")
}

// Regions returns every slide and part region in the module's Go files.
func Regions() ([]Region, error) {
	files, err := GoFiles()
	if err != nil {
		return nil, err
	}
	var found []Region
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(Root, rel))
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			var r Region
			var indent, end string
			if m := slideMarker.FindStringSubmatch(line); m != nil {
				indent, r.Segment, r.Anchor, end = m[1], m[2], m[3], "// end slide"
			} else if m := partMarker.FindStringSubmatch(line); m != nil {
				indent, r.Segment, r.Anchor, end = m[1], m[2], m[3], "// end part"
				r.Part, _ = strconv.Atoi(m[4])
			} else {
				continue
			}
			stop := -1
			for j := i + 1; j < len(lines); j++ {
				if lines[j] == indent+end {
					stop = j
					break
				}
			}
			if stop < 0 {
				return nil, fmt.Errorf("%s:%d has a marker with no %q", rel, i+1, end)
			}
			r.File, r.Line = rel, i+1
			r.Code = strings.TrimRight(dedent(lines[i+1:stop]), "\n")
			found = append(found, r)
		}
	}
	return found, nil
}

// SlideCode returns the first Go block after anchor in a segment.
func SlideCode(segment, anchor string) (string, error) {
	data, err := os.ReadFile(filepath.Join(Segments, segment))
	if err != nil {
		return "", err
	}
	text := string(data)
	start := strings.Index(text, anchor)
	if start < 0 {
		return "", fmt.Errorf("%s has no %q", segment, anchor)
	}
	m := Fence.FindStringSubmatch(text[start:])
	if m == nil {
		return "", fmt.Errorf("%s has no Go block after %q", segment, anchor)
	}
	return strings.TrimRight(m[1], "\n"), nil
}

// Block is one Go code block on a slide.
type Block struct {
	Segment string
	Line    int
	Code    string
}

// AllBlocks returns every Go block in the Deck's segments.
func AllBlocks() ([]Block, error) {
	segs, err := SegmentFiles()
	if err != nil {
		return nil, err
	}
	var out []Block
	for _, s := range segs {
		data, err := os.ReadFile(filepath.Join(Segments, s))
		if err != nil {
			return nil, err
		}
		text := string(data)
		for _, m := range Fence.FindAllStringSubmatchIndex(text, -1) {
			out = append(out, Block{
				Segment: s,
				Line:    strings.Count(text[:m[0]], "\n") + 1,
				Code:    text[m[2]:m[3]],
			})
		}
	}
	return out, nil
}
