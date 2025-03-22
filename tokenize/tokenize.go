package tokenize

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"unsafe"
)

type Pair struct {
	L rune
	R rune
}

func (pair Pair) Equals(another Pair) bool {
	return pair.L == another.L && pair.R == another.R
}

func RenderTokens(pairs []Pair, tokens []rune) {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		pair := pairs[token]

		if pair.L == token {
			fmt.Printf("%c", token)
		} else {
			fmt.Printf("[%U]", token)
		}
	}
	fmt.Printf("\n")
}

func DumpPairs(pairs []Pair, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal("Could not open file")
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, pairs)
	if err != nil {
		log.Fatal("Write failed")
	}
}

func LoadPairs(filePath string) []Pair {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Could not open file")
	}
	defer file.Close()

	bin, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Read failed")
	}
	binReader := bytes.NewReader(bin)

	pairCount := len(bin) / int(unsafe.Sizeof(Pair{}))
	pairs := make([]Pair, pairCount)

	err = binary.Read(binReader, binary.LittleEndian, &pairs)
	if err != nil {
		log.Fatal("Failed to read bytes: ", err)
	}

	return pairs
}

func Tokenize(text string) []Pair {
	freq := make(map[Pair]int)
	tokensIn := []rune{}
	tokensOut := []rune{}
	pairs := []Pair{}

	for i := 0; i < 256; i++ {
		pairs = append(pairs, Pair{
			L: rune(i),
		})
	}

	for _, r := range text {
		tokensIn = append(tokensIn, r)
	}

	for {
		for i := 0; i < len(tokensIn)-1; i++ {
			key := Pair{
				L: tokensIn[i],
				R: tokensIn[i+1],
			}
			freq[key]++
		}

		var best_pair Pair
		for pair, count := range freq {
			if count > freq[best_pair] {
				best_pair = pair
			}
		}

		if freq[best_pair] <= 1 {
			break
		}

		pairs = append(pairs, best_pair)

		for i := 0; i < len(tokensIn); i++ {
			if i+1 >= len(tokensIn) {
				tokensOut = append(tokensOut, tokensIn[i])
				break
			}

			pair := Pair{
				L: tokensIn[i],
				R: tokensIn[i+1],
			}

			if pair.Equals(best_pair) {
				tokensOut = append(tokensOut, rune(len(pairs)-1))
				i++
			} else {
				tokensOut = append(tokensOut, tokensIn[i])
			}
		}
		tokensIn = tokensOut
		tokensOut = nil
		clear(freq)
	}

	return pairs
}
