package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

// errInvalidFlagValue signals a usage error distinct from flag.ErrHelp,
// so run() maps it to exit code 2.
var errInvalidFlagValue = errors.New("invalid flag value")

// letterFreq maps a letter to its occurrence count.
type letterFreq map[rune]int64

// config holds the resolved command-line options.
type config struct {
	caseSensitive bool
	sortBy        string // "count" or "alpha"
	graph         bool
}

// graphChar fills graph bars.
const graphChar = '■'

// maxBarLen is the bar length in characters for the most frequent letter.
const maxBarLen = 20

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run contains all program logic and returns an exit code.
//
// Exit codes:
//
// 0 - success, every input processed cleanly
// 1 - one or more inputs failed (e.g. file not found); others still processed
// 2 - usage error (bad flags)
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg, fileArgs, err := parseFlags(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	freq := make(letterFreq, 64)
	var total int64
	hadError := false

	if len(fileArgs) == 0 {
		fileArgs = []string{"-"} // no args: read stdin
	}

	for _, name := range fileArgs {
		n, err := accumulate(name, stdin, cfg.caseSensitive, freq)
		total += n
		if err != nil {
			fmt.Fprintf(stderr, "freaks: %v\n", err)
			hadError = true
		}
	}

	printFreq(stdout, freq, total, cfg.sortBy, cfg.graph)

	if hadError {
		return 1
	}
	return 0
}

// parseFlags defines and parses the CLI flags, writing usage output to stderr on error.
func parseFlags(args []string, stderr io.Writer) (config, []string, error) {
	fs := flag.NewFlagSet("freaks", flag.ContinueOnError)
	fs.SetOutput(stderr)

	caseSensitive := fs.Bool("case-sensitive", false,
		"count uppercase and lowercase letters separately (default: fold to lowercase)")
	sortBy := fs.String("sort", "count",
		`order results by "count" (most frequent first) or "alpha" (alphabetical)`)
	graph := fs.Bool("graph", false,
		"append a relative-scale ASCII bar (1-20 chars) after the count")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: freaks [flags] [file...]\n")
		fmt.Fprintf(stderr, "Reports letter frequency for the given files, or stdin if none are given.\n")
		fmt.Fprintf(stderr, "Use \"-\" to read stdin explicitly (e.g. alongside other files).\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return config{}, nil, err
	}

	if *sortBy != "count" && *sortBy != "alpha" {
		fmt.Fprintf(stderr, "freaks: invalid -sort value %q (want \"count\" or \"alpha\")\n", *sortBy)
		fs.Usage()
		return config{}, nil, errInvalidFlagValue
	}

	return config{caseSensitive: *caseSensitive, sortBy: *sortBy, graph: *graph}, fs.Args(), nil
}

// accumulate opens name (or stdin, for "-") and folds its letter counts into freq.
// It returns the number of letters counted from this input.
func accumulate(name string, stdin io.Reader, caseSensitive bool, freq letterFreq) (int64, error) {
	if name == "-" {
		n, err := countLetters(stdin, caseSensitive, freq)
		if err != nil {
			return n, fmt.Errorf("standard input: %w", err)
		}
		return n, nil
	}

	f, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	defer func() { _ = f.Close() }()

	n, err := countLetters(f, caseSensitive, freq)
	if err != nil {
		return n, fmt.Errorf("%s: %w", name, err)
	}
	return n, nil
}

// countLetters streams r rune-by-rune, tallying letters into freq.
func countLetters(r io.Reader, caseSensitive bool, freq letterFreq) (int64, error) {
	br := bufio.NewReaderSize(r, 64*1024)
	var n int64

	for {
		ru, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			return n, err
		}

		if !unicode.IsLetter(ru) {
			continue
		}
		if !caseSensitive {
			ru = unicode.ToLower(ru)
		}
		freq[ru]++
		n++
	}
}

// printFreq writes one line per letter.
func printFreq(w io.Writer, freq letterFreq, total int64, sortBy string, graph bool) {
	letters := make([]rune, 0, len(freq))
	var maxCount int64
	for r, count := range freq {
		letters = append(letters, r)
		if count > maxCount {
			maxCount = count
		}
	}

	switch sortBy {
	case "alpha":
		sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })
	default: // "count"
		sort.Slice(letters, func(i, j int) bool {
			if freq[letters[i]] != freq[letters[j]] {
				return freq[letters[i]] > freq[letters[j]] // descending
			}
			return letters[i] < letters[j] // stable tie-break
		})
	}

	for _, v := range letters {
		count := freq[v]
		var pct float64
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}

		if !graph {
			fmt.Fprintf(w, "%c\t%.2f%%\t%d\n", v, pct, count)
			continue
		}
		fmt.Fprintf(w, "%c\t%.2f%%\t%d\t%s\n", v, pct, count, bar(count, maxCount))
	}
}

// bar renders graph relative to maxCount as a 1-20 character bar
func bar(count, maxCount int64) string {
	n := maxBarLen
	if maxCount > 0 && count != maxCount {
		n = int(float64(count)/float64(maxCount)*maxBarLen + 0.5) // round to nearest
	}
	if n < 1 {
		n = 1
	}
	if n > maxBarLen {
		n = maxBarLen
	}
	return strings.Repeat(string(graphChar), n)
}
