package service

import (
	"fmt"
	"strings"
)

// Diffing two revisions of an entry.
//
// Line-level LCS with a word-level pass inside a line that was replaced. Prose
// is what these documents are: a line diff alone reports "this paragraph is
// gone, this other paragraph appeared" for a one-word correction, which tells a
// reader nothing about what actually changed.
//
// No dependency for this. The alternative is a diff library pulled in for one
// screen, in a project whose whole distribution story is a single CGo-free
// binary — and the algorithm below is the textbook one.

// DiffOp is what happened to a line or a word.
type DiffOp string

const (
	DiffEqual  DiffOp = "equal"
	DiffAdd    DiffOp = "add"
	DiffRemove DiffOp = "remove"
)

// DiffWord is one token inside a replaced line.
type DiffWord struct {
	Op   DiffOp `json:"op"`
	Text string `json:"text"`
}

// DiffLine is one line of the comparison. Words is filled only where a line was
// replaced by a recognizably similar one.
type DiffLine struct {
	Op    DiffOp     `json:"op"`
	Text  string     `json:"text"`
	Words []DiffWord `json:"words,omitempty"`
}

// DiffResult is the whole comparison, in both the form a UI renders and the form
// a terminal reads.
type DiffResult struct {
	Lines     []DiffLine `json:"lines"`
	Added     int        `json:"added"`
	Removed   int        `json:"removed"`
	Unchanged int        `json:"unchanged"`
	Unified   string     `json:"unified"`
	// Truncated is set when the documents were too large to align line by line
	// and the comparison fell back to "everything was replaced".
	Truncated bool `json:"truncated,omitempty"`
}

// maxDiffLines caps the O(n*m) alignment table. Two 4000-line documents would
// allocate 16M cells; entries are articles, not log files, so the cap is far
// above anything real and exists to keep a pathological input from taking the
// process down with it.
const maxDiffLines = 4000

// DiffText compares two documents line by line.
func DiffText(before, after string) DiffResult {
	oldLines := splitLines(before)
	newLines := splitLines(after)

	if len(oldLines) > maxDiffLines || len(newLines) > maxDiffLines {
		// Report the shape of the change, not the change itself. Emitting every
		// line of both documents as remove-then-add is technically a diff and
		// practically useless: nothing collapses (every line differs), the
		// reader gets eight thousand rows, and the payload doubles the size of
		// two documents nobody asked to see twice. The counts and the flag are
		// what a caller can act on.
		return DiffResult{
			Truncated: true,
			Removed:   len(oldLines),
			Added:     len(newLines),
			Lines:     []DiffLine{},
		}
	}

	lines := alignLines(oldLines, newLines)
	lines = pairReplacements(lines)

	res := DiffResult{Lines: lines}
	for _, l := range lines {
		switch l.Op {
		case DiffAdd:
			res.Added++
		case DiffRemove:
			res.Removed++
		default:
			res.Unchanged++
		}
	}
	res.Unified = buildUnified(lines)
	return res
}

// splitLines keeps empty lines: in markdown a blank line is a paragraph break,
// and dropping it would report structural changes that did not happen.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// alignLines is Myers-style LCS via the classic dynamic programming table,
// walked backwards into a linear sequence of operations.
func alignLines(a, b []string) []DiffLine {
	n, m := len(a), len(b)
	// lcs[i][j] = length of the longest common subsequence of a[i:] and b[j:]
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	out := make([]DiffLine, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, DiffLine{Op: DiffEqual, Text: a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, DiffLine{Op: DiffRemove, Text: a[i]})
			i++
		default:
			out = append(out, DiffLine{Op: DiffAdd, Text: b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, DiffLine{Op: DiffRemove, Text: a[i]})
	}
	for ; j < m; j++ {
		out = append(out, DiffLine{Op: DiffAdd, Text: b[j]})
	}
	return out
}

// pairReplacements finds a removed line immediately followed by an added one and,
// when they are similar enough to be the same sentence edited, attaches a
// word-level breakdown to both.
//
// The similarity gate matters: without it, an unrelated paragraph deleted next
// to an unrelated paragraph added would render as a soup of highlighted words
// that reads as noise.
func pairReplacements(lines []DiffLine) []DiffLine {
	for i := 0; i+1 < len(lines); i++ {
		if lines[i].Op != DiffRemove || lines[i+1].Op != DiffAdd {
			continue
		}
		before, after := lines[i].Text, lines[i+1].Text
		if !similarEnough(before, after) {
			continue
		}
		words := diffWords(before, after)
		lines[i].Words = keepSide(words, DiffRemove)
		lines[i+1].Words = keepSide(words, DiffAdd)
		i++ // both lines of the pair are handled
	}
	return lines
}

// similarEnough is a cheap token-overlap ratio. 0.4 was chosen so that a
// sentence with a couple of words changed pairs up, while two different
// sentences that happen to share "the" and "and" do not.
func similarEnough(a, b string) bool {
	aw, bw := strings.Fields(a), strings.Fields(b)
	if len(aw) == 0 || len(bw) == 0 {
		return false
	}
	counts := make(map[string]int, len(aw))
	for _, w := range aw {
		counts[w]++
	}
	shared := 0
	for _, w := range bw {
		if counts[w] > 0 {
			counts[w]--
			shared++
		}
	}
	longer := len(aw)
	if len(bw) > longer {
		longer = len(bw)
	}
	return float64(shared)/float64(longer) >= 0.4
}

// diffWords aligns two lines token by token, keeping the whitespace that
// followed each token so the rendered line reads as it was written.
func diffWords(a, b string) []DiffWord {
	at, bt := tokenize(a), tokenize(b)
	n, m := len(at), len(bt)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if at[i] == bt[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	out := make([]DiffWord, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case at[i] == bt[j]:
			out = append(out, DiffWord{Op: DiffEqual, Text: at[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, DiffWord{Op: DiffRemove, Text: at[i]})
			i++
		default:
			out = append(out, DiffWord{Op: DiffAdd, Text: bt[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, DiffWord{Op: DiffRemove, Text: at[i]})
	}
	for ; j < m; j++ {
		out = append(out, DiffWord{Op: DiffAdd, Text: bt[j]})
	}
	return out
}

// tokenize splits on spaces while keeping the separator attached to the token
// before it, so joining the tokens reproduces the line exactly.
func tokenize(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			out = append(out, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func keepSide(words []DiffWord, side DiffOp) []DiffWord {
	out := make([]DiffWord, 0, len(words))
	for _, w := range words {
		if w.Op == DiffEqual || w.Op == side {
			out = append(out, w)
		}
	}
	return out
}

// buildUnified renders the classic +/- form with three lines of context, for
// callers that want text: an MCP tool result, a terminal, a commit-style view.
func buildUnified(lines []DiffLine) string {
	const context = 3
	keep := make([]bool, len(lines))
	for i, l := range lines {
		if l.Op == DiffEqual {
			continue
		}
		for j := i - context; j <= i+context; j++ {
			if j >= 0 && j < len(lines) {
				keep[j] = true
			}
		}
	}

	var b strings.Builder
	gap := false
	for i, l := range lines {
		if !keep[i] {
			gap = true
			continue
		}
		if gap && b.Len() > 0 {
			b.WriteString("@@\n")
		}
		gap = false
		switch l.Op {
		case DiffAdd:
			b.WriteString("+ ")
		case DiffRemove:
			b.WriteString("- ")
		default:
			b.WriteString("  ")
		}
		b.WriteString(l.Text)
		b.WriteString("\n")
	}
	return b.String()
}

// DiffSummary is the one-line form: "+12 −3".
func DiffSummary(d DiffResult) string {
	return fmt.Sprintf("+%d −%d", d.Added, d.Removed)
}
