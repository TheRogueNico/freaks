package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

func createGraph(data map[rune]int) string {
	max := 0
	for _, n := range data {
		if n > max {
			max = n
		}
	}

	graph := make([]string, 0, len(data))
	for k, n := range data {
		var character string
		switch k {
		case '\n':
			character = `'\n'`
		case '\t':
			character = `'\t'`
		case ' ':
			character = "' '"
		default:
			character = string(k)
		}

		barLen := int(math.Round(float64(n) / float64(max) * 40))
		bar := strings.Repeat("█", barLen)
		graph = append(graph, fmt.Sprintf("[%-4s:%3d] ▌%s", character, n, bar))
	}
	return strings.Join(graph, "\n")
}

func main() {
	fmt.Print("Enter input (Ctrl-D to send & exit):\n> ")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
	input := string(data)
	fmt.Println(input)

	freq := make(map[rune]int)
	for _, v := range input {
		freq[v]++
	}

	fmt.Println(createGraph(freq))
}
