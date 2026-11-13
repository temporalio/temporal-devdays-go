// ABOUTME: Skip. Finishes named TODOs in the Exercise by copying them from the
// Solution.
//
//	workshop-solve <exercise-dir> <solution-dir> <file> "TODO 2a" ["TODO 2b" ...]
//
// Replaces the lines between `// TODO N start` and `// TODO N end` in
// <exercise-dir>/<file> with the same region from <solution-dir>/<file>, and
// nothing else, so a learner who skips one Challenge keeps their own work on
// every other TODO. The learner's file is kept as <file>.bak the first time.
// It only splices; it never overwrites a whole file.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func bounds(lines []string, todo, file string) (int, int) {
	q := regexp.QuoteMeta(todo)
	startRe := regexp.MustCompile(`//\s*` + q + ` start\b`)
	endRe := regexp.MustCompile(`//\s*` + q + ` end\b`)
	start, end := -1, -1
	for i, l := range lines {
		if start < 0 && startRe.MatchString(l) {
			start = i
		} else if start >= 0 && end < 0 && endRe.MatchString(l) {
			end = i
		}
	}
	if start < 0 || end < 0 {
		fmt.Fprintf(os.Stderr, "workshop-solve: %s has no '// %s start' / '// %s end' pair\n", file, todo, todo)
		os.Exit(1)
	}
	return start, end
}

func main() {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, `usage: workshop-solve <exercise-dir> <solution-dir> <file> "TODO N" ["TODO M" ...]`)
		os.Exit(2)
	}
	exercise, solution, rel, todos := os.Args[1], os.Args[2], os.Args[3], os.Args[4:]
	target := filepath.Join(exercise, rel)
	source := filepath.Join(solution, rel)
	orig, err := os.ReadFile(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "workshop-solve:", err)
		os.Exit(1)
	}
	refBytes, err := os.ReadFile(source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "workshop-solve:", err)
		os.Exit(1)
	}
	lines := strings.Split(string(orig), "\n")
	ref := strings.Split(string(refBytes), "\n")
	for _, todo := range todos {
		s, e := bounds(lines, todo, target)
		rs, re := bounds(ref, todo, source)
		next := append([]string{}, lines[:s]...)
		next = append(next, ref[rs:re+1]...)
		lines = append(next, lines[e+1:]...)
	}
	backup := target + ".bak"
	if _, err := os.Stat(backup); os.IsNotExist(err) {
		_ = os.WriteFile(backup, orig, 0o644)
	}
	if err := os.WriteFile(target, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "workshop-solve:", err)
		os.Exit(1)
	}
	fmt.Printf("workshop-solve: finished %s in %s\n", strings.Join(todos, ", "), target)
}
