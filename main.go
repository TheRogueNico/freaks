package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func createGraph(data map[rune]int) string {
	graph := make([]string, 0)
	character := ""
	for k, n := range data {
		switch k {
		case '\n':
			character = "'\\n'"
		case '\t':
			character = "'\\t'"
		case ' ':
			character = "' '"
		default:
			character = string(k)
		}
		bar := strings.Repeat("█", n)
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
