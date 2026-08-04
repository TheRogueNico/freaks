package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func createGraph(data map[rune]int) string {
	graph := make([]string, 0)
	for k, n := range data {
		if k == '\n' {
			continue
		}
		bar := strings.Repeat("█", n)
		graph = append(graph, fmt.Sprintf("[%2c : %-3d]: ▌%s", k, n, bar))
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
