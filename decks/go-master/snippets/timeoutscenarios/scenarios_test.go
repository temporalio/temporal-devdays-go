// Package timeoutscenarios checks that every TimeoutSequence slide steps
// through exactly its scenario's captions.
//
// A slide's `clicks:` sets how many times the presenter can advance it. If it
// is lower than the scenario's steps, the last ones are unreachable; if higher,
// the diagram sits frozen on its final caption for extra clicks. Nothing in
// Slidev checks this, so this does, reading the scenarios straight out of
// components/timeout-scenarios.ts.
package timeoutscenarios

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/deck"
)

type scenario struct {
	rows    int
	ats     []int // every click something appears or closes on
	windows [][2]int
	say     int
}

var (
	header    = regexp.MustCompile(`^  '?([\w-]+)'?: \{$`)
	atRe      = regexp.MustCompile(`\bat: (\d+)`)
	fromTo    = regexp.MustCompile(`from: (\d+), to: (\d+)`)
	slideRe   = regexp.MustCompile(`(?m)^---\n(?:layout:|src:)`)
	useRe     = regexp.MustCompile(`<TimeoutSequence scenario="([\w-]+)"`)
	clicksRe  = regexp.MustCompile(`(?m)^clicks: (\d+)$`)
	sayString = regexp.MustCompile(`^\s+(['"]).*(['"]),?$`)
)

func scenarios(t *testing.T) map[string]*scenario {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(deck.Root, "..", "components", "timeout-scenarios.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, body, ok := strings.Cut(string(data), "export const SCENARIOS")
	if !ok {
		t.Fatal("timeout-scenarios.ts has no SCENARIOS")
	}
	found := map[string]*scenario{}
	var cur *scenario
	section := ""
	for _, line := range strings.Split(body, "\n") {
		if m := header.FindStringSubmatch(line); m != nil {
			cur = &scenario{}
			found[m[1]] = cur
			continue
		}
		if cur == nil {
			continue
		}
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "rows: ["):
			section = "rows"
			continue
		case strings.HasPrefix(trimmed, "windows: ["):
			section = "windows"
			if strings.HasSuffix(trimmed, "[],") {
				section = ""
			}
			continue
		case strings.HasPrefix(trimmed, "events: ["):
			section = "events"
			continue
		case strings.HasPrefix(trimmed, "say: ["):
			section = "say"
			continue
		case trimmed == "],":
			section = ""
			continue
		}
		switch section {
		case "rows", "windows", "events":
			if !strings.HasPrefix(trimmed, "{") {
				continue
			}
			for _, m := range atRe.FindAllStringSubmatch(line, -1) {
				n, _ := strconv.Atoi(m[1])
				cur.ats = append(cur.ats, n)
			}
			if section == "rows" {
				cur.rows++
			}
			if section == "windows" {
				m := fromTo.FindStringSubmatch(line)
				a, _ := strconv.Atoi(m[1])
				b, _ := strconv.Atoi(m[2])
				cur.windows = append(cur.windows, [2]int{a, b})
			}
		case "say":
			if sayString.MatchString(line) {
				cur.say++
			}
		}
	}
	return found
}

type use struct {
	segment, name string
	clicks        int
}

func uses(t *testing.T) []use {
	t.Helper()
	segs, err := deck.SegmentFiles()
	if err != nil {
		t.Fatal(err)
	}
	var found []use
	for _, seg := range segs {
		data, err := os.ReadFile(filepath.Join(deck.Segments, seg))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		idx := slideRe.FindAllStringIndex(text, -1)
		for k, loc := range idx {
			end := len(text)
			if k+1 < len(idx) {
				end = idx[k+1][0]
			}
			slide := text[loc[0]:end]
			m := useRe.FindStringSubmatch(slide)
			if m == nil {
				continue
			}
			clicks := -1
			if c := clicksRe.FindStringSubmatch(slide); c != nil {
				clicks, _ = strconv.Atoi(c[1])
			}
			found = append(found, use{seg, m[1], clicks})
		}
	}
	return found
}

func TestEveryScenarioIsOnASlide(t *testing.T) {
	s := scenarios(t)
	if len(s) < 7 {
		t.Fatalf("found %d scenarios, want at least 7", len(s))
	}
	var want, got []string
	for name := range s {
		want = append(want, name)
	}
	seen := map[string]bool{}
	for _, u := range uses(t) {
		if !seen[u.name] {
			got = append(got, u.name)
			seen[u.name] = true
		}
	}
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, ",") != strings.Join(got, ",") {
		t.Fatalf("scenarios %v, but slides use %v", want, got)
	}
}

func TestEverySlideSetsClicksToItsScenarioSteps(t *testing.T) {
	s := scenarios(t)
	for _, u := range uses(t) {
		t.Run(u.segment+":"+u.name, func(t *testing.T) {
			sc, ok := s[u.name]
			if !ok {
				t.Fatalf("%s uses unknown scenario %s", u.segment, u.name)
			}
			if u.clicks != sc.say-1 {
				t.Fatalf("%s: clicks: %d, but %s has %d captions", u.segment, u.clicks, u.name, sc.say)
			}
		})
	}
}

func TestEveryScenarioIsInternallyConsistent(t *testing.T) {
	for name, sc := range scenarios(t) {
		t.Run(name, func(t *testing.T) {
			if sc.say == 0 || sc.rows == 0 {
				t.Fatalf("parsed %d captions and %d rows; did the file's layout change?", sc.say, sc.rows)
			}
			for _, at := range sc.ats {
				if at > sc.say-1 {
					t.Errorf("something appears on click %d, which has no caption (%d captions)", at, sc.say)
				}
			}
			for _, w := range sc.windows {
				if !(0 <= w[0] && w[0] <= w[1] && w[1] < sc.rows) {
					t.Errorf("window rows %d-%d outside the %d rows", w[0], w[1], sc.rows)
				}
			}
		})
	}
}
