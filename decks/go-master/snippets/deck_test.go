package snippets

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/deck"
)

// A slide and the package that runs its code must say exactly the same thing.
func TestSlidesMatch(t *testing.T) {
	regions, err := deck.Regions()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, r := range regions {
		if r.Part != 0 {
			continue
		}
		n++
		t.Run(r.Segment+" "+r.Anchor, func(t *testing.T) {
			slide, err := deck.SlideCode(r.Segment, r.Anchor)
			if err != nil {
				t.Fatal(err)
			}
			if r.Code != slide {
				t.Errorf("%s:%d has drifted from %s (%s). Edit the slide and the snippet together.\n%s",
					r.File, r.Line, r.Segment, r.Anchor, diff(slide, r.Code))
			}
		})
	}
	if n < 45 {
		t.Errorf("only %d slide regions; every code slide should have one", n)
	}
}

// A slide that shows several files in one block must equal its parts, in
// order, joined by one blank line.
func TestPartsMatch(t *testing.T) {
	regions, err := deck.Regions()
	if err != nil {
		t.Fatal(err)
	}
	groups := map[string][]deck.Region{}
	var keys []string
	for _, r := range regions {
		if r.Part == 0 {
			continue
		}
		k := r.Segment + " " + r.Anchor
		if _, ok := groups[k]; !ok {
			keys = append(keys, k)
		}
		groups[k] = append(groups[k], r)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts := groups[k]
		t.Run(k, func(t *testing.T) {
			sort.Slice(parts, func(i, j int) bool { return parts[i].Part < parts[j].Part })
			var codes []string
			for i, p := range parts {
				if p.Part != i+1 {
					t.Fatalf("parts of %s are numbered %d at position %d; number them 1, 2, 3...", k, p.Part, i+1)
				}
				codes = append(codes, p.Code)
			}
			slide, err := deck.SlideCode(parts[0].Segment, parts[0].Anchor)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(codes, "\n\n"); got != slide {
				t.Errorf("the parts for %s have drifted from the slide. Edit them together.\n%s", k, diff(slide, got))
			}
		})
	}
}

// Every Go block in the Deck parses, fragments included. A slide may show a
// whole file, top-level declarations without their package clause, a function
// body, the fields of a composite literal or struct, or a `case` clause; try it
// at each level it could plausibly sit at.
func TestSlidesParse(t *testing.T) {
	blocks, err := deck.AllBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) < 60 {
		t.Errorf("the Deck has only %d Go blocks", len(blocks))
	}
	for _, b := range blocks {
		t.Run(fmt.Sprintf("%s:%d", b.Segment, b.Line), func(t *testing.T) {
			if !parses(b.Code) {
				t.Errorf("%s:%d is not valid Go at any level:\n%s", b.Segment, b.Line, firstError(b.Code))
			}
		})
	}
}

func wrappers(code string) []string {
	return []string{
		code,
		"package p\n" + code,
		"package p\nfunc _() {\n" + code + "\n}",
		"package p\nvar _ = T{\n" + code + "\n}",
		"package p\ntype _ struct {\n" + code + "\n}",
		"package p\nfunc _() {\nswitch {\n" + code + "\n}\n}",
		"package p\nfunc _() {\nselect {\n" + code + "\n}\n}",
	}
}

func parses(code string) bool {
	for _, w := range wrappers(code) {
		if _, err := parser.ParseFile(token.NewFileSet(), "slide.go", w, parser.SkipObjectResolution); err == nil {
			return true
		}
	}
	return false
}

func firstError(code string) string {
	_, err := parser.ParseFile(token.NewFileSet(), "slide.go", "package p\n"+code, 0)
	return fmt.Sprint(err)
}

// Every snippet file is gofmt-clean, so the code on the slides is too.
func TestGofmt(t *testing.T) {
	files, err := deck.GoFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(deck.Root, rel))
		if err != nil {
			t.Fatal(err)
		}
		out, err := format.Source(src)
		if err != nil {
			t.Errorf("%s: %v", rel, err)
			continue
		}
		if string(out) != string(src) {
			t.Errorf("%s is not gofmt-clean; run gofmt -w", rel)
		}
	}
}

// Every code-walkthrough bubble must point at the code it talks about.
//
// A walkthrough pairs a `{all|3|5-6|...}` highlight sequence with a `notes:`
// list, one bubble per step. Nothing in Slidev checks that the two agree, so an
// edit that shifts the code by a line leaves every bubble aimed at the wrong
// line. This reads the segments statically; scripts/check-walkthroughs.mjs
// checks the rendered positions in a browser.
func TestWalkthroughs(t *testing.T) {
	ws, err := walkthroughs()
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) < 40 {
		t.Errorf("the Deck has only %d walkthroughs", len(ws))
	}
	for _, w := range ws {
		t.Run(w.id, func(t *testing.T) {
			if len(w.notes) != len(w.steps) {
				t.Fatalf("%d highlight steps but %d notes: each bubble would land on the wrong step", len(w.steps), len(w.notes))
			}
			for i, spec := range w.steps {
				lines := parseRanges(spec)
				maxLine := 0
				for _, n := range lines {
					maxLine = max(maxLine, n)
				}
				if maxLine > len(w.code) || len(lines) == 0 {
					t.Errorf("step %d highlights line %d of %d", i+1, maxLine, len(w.code))
					continue
				}
				lit := make([]string, len(lines))
				for j, n := range lines {
					lit[j] = w.code[n-1]
				}
				// A range that begins or ends on a blank line is the signature
				// of code that moved under a highlight.
				if strings.TrimSpace(lit[0]) == "" || strings.TrimSpace(lit[len(lit)-1]) == "" {
					t.Errorf("step %d: lines %s start or end blank; did the code shift? Lit: %q", i+1, spec, lit)
				}
				joined := strings.Join(lit, "\n")
				for _, m := range codeTag.FindAllStringSubmatch(w.notes[i], -1) {
					token := strings.TrimRight(unescape(m[1]), "()=:;")
					if !strings.Contains(joined, token) {
						t.Errorf("step %d: note mentions <code>%s</code> but the lit lines are %q", i+1, m[1], lit)
					}
				}
			}
		})
	}
}

type walkthrough struct {
	id    string
	steps []string
	notes []string
	code  []string
}

var (
	slideSplit = regexp.MustCompile(`(?m)^---\n(?:layout:|src:)`)
	codeTag    = regexp.MustCompile(`<code>(.*?)</code>`)
	headingRe  = regexp.MustCompile(`(?m)^heading: (.+)$`)
	walkFence  = regexp.MustCompile("(?ms)^```(?:go|txt) \\{([^}]*)\\}\\n(.*?)^```$")
)

func walkthroughs() ([]walkthrough, error) {
	segs, err := deck.SegmentFiles()
	if err != nil {
		return nil, err
	}
	var found []walkthrough
	for _, seg := range segs {
		data, err := os.ReadFile(filepath.Join(deck.Segments, seg))
		if err != nil {
			return nil, err
		}
		text := string(data)
		idx := slideSplit.FindAllStringIndex(text, -1)
		for k, loc := range idx {
			// Keep the `layout:` the split consumed: start after the `---\n`.
			end := len(text)
			if k+1 < len(idx) {
				end = idx[k+1][0]
			}
			slide := text[loc[0]+4 : end]
			sep := strings.Index(slide, "\n---\n")
			if sep < 0 {
				continue
			}
			front, body := slide[:sep], slide[sep+5:]
			if !strings.Contains(front, "layout: code-walkthrough") {
				continue
			}
			hm := headingRe.FindStringSubmatch(front)
			if hm == nil {
				return nil, fmt.Errorf("%s: a code-walkthrough has no heading", seg)
			}
			heading := strings.Trim(hm[1], `"`)
			fm := walkFence.FindStringSubmatch(body)
			if fm == nil {
				return nil, fmt.Errorf("%s: %s has no highlighted Go or txt block", seg, heading)
			}
			specs := strings.Split(fm[1], "|")
			if specs[0] == "all" { // `all` burns click 0, unannotated.
				specs = specs[1:]
			}
			code := strings.Split(fm[2], "\n")
			found = append(found, walkthrough{
				id:    seg + ": " + heading,
				steps: specs,
				notes: parseNotes(front),
				code:  code[:len(code)-1],
			})
		}
	}
	return found, nil
}

func parseRanges(spec string) []int {
	var out []int
	for _, part := range strings.Split(spec, ",") {
		a, b, ok := strings.Cut(part, "-")
		lo, _ := strconv.Atoi(strings.TrimSpace(a))
		hi := lo
		if ok {
			hi, _ = strconv.Atoi(strings.TrimSpace(b))
		}
		for n := lo; n <= hi; n++ {
			out = append(out, n)
		}
	}
	return out
}

func parseNotes(front string) []string {
	_, after, ok := strings.Cut(front, "notes:")
	if !ok {
		return nil
	}
	var notes []string
	for _, line := range strings.Split(after, "\n")[1:] {
		if !strings.HasPrefix(line, "  - ") {
			break
		}
		note := strings.TrimSpace(line[4:])
		if len(note) >= 2 && note[0] == '"' && note[len(note)-1] == '"' {
			note = note[1 : len(note)-1]
		}
		notes = append(notes, note)
	}
	return notes
}

func unescape(s string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&amp;", "&").Replace(s)
}

// diff shows the first line where want and got part ways.
func diff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < max(len(w), len(g)); i++ {
		var a, b string
		if i < len(w) {
			a = w[i]
		}
		if i < len(g) {
			b = g[i]
		}
		if a != b || i >= len(w) || i >= len(g) {
			return fmt.Sprintf("first difference at line %d:\n  slide:   %q\n  snippet: %q", i+1, a, b)
		}
	}
	return ""
}
