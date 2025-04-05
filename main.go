package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"tokenizer/tokenize"
)

func renderToken(pairs []tokenize.Pair, token rune, sb *strings.Builder) {
	if token == pairs[token].L {
		sb.WriteString(string(token))
		return
	}

	renderToken(pairs, pairs[token].L, sb)
	renderToken(pairs, pairs[token].R, sb)
}

func inspect(pairs []tokenize.Pair) {
	var sb strings.Builder
	for token := 1; token < len(pairs); token++ {
		renderToken(pairs, rune(token), &sb)

		fmt.Printf("%U => {%s}\n", token, sb.String())
		sb.Reset()
	}
}

func main() {
	args := os.Args[1:]

	mode := args[0]
	switch mode {
	case "token":
		{
			filePath := args[1]
			if filePath == "" {
				log.Fatal("Missing source file argument")
			}

			destinationPath := args[2]
			if filePath == "" {
				log.Fatal("Missing destination file argument")
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatal("Failed to read file:", err)
			}

			text := string(content[:])
			pairs := tokenize.Tokenize(text)

			tokenize.DumpPairs(pairs, destinationPath)
			fmt.Printf("Saved pairs")
		}
	case "inspect":
		{
			filePath := args[1]
			if filePath == "" {
				log.Fatal("Missing binary source file argument")
			}

			pairs := tokenize.LoadPairs(filePath)
			fmt.Printf("Loaded: %d pairs\n", len(pairs))
			inspect(pairs)
		}
	default:
		fmt.Println("Usage: <mode> <parameters>")
		fmt.Println("Example: token source.txt output.bin")
		fmt.Println("Example: inspect output.bin")
	}
}
