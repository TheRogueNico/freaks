package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"unicode"
)

// letterFreq counts occurrences of each letter.
type letterFreq map[rune]int64

// process streams r and prints letter frequencies.
func process(r io.Reader) error {
	freq, err := countLetters(r)
	if err != nil {
		return err
	}
	printFreq(freq)
	return nil
}

// countLetters reads one rune at a time via bufio.Reader,
// so memory use stays constant regardless of input size.
func countLetters(r io.Reader) (letterFreq, error) {
	br := bufio.NewReaderSize(r, 64*1024)
	freq := make(letterFreq, 64)

	for {
		ru, _, err := br.ReadRune()
		if err != nil {
			if err == io.EOF {
				return freq, nil
			}
			return nil, err
		}

		if !unicode.IsLetter(ru) {
			continue // skip whitespace, digits, punctuation, etc.
		}
		freq[unicode.ToLower(ru)]++
	}
}

// printFreq prints results sorted by letter.
func printFreq(freq letterFreq) {
	letters := make([]rune, 0, len(freq))
	for r := range freq {
		letters = append(letters, r)
	}
	sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })

	for _, r := range letters {
		fmt.Printf("%c: %d\n", r, freq[r])
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return process(os.Stdin)
	}

	var firstErr error
	for _, name := range args {
		if err := processArg(name); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func processArg(name string) error {
	if name == "-" {
		return process(os.Stdin)
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return process(f)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
