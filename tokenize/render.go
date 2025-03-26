package tokenize

import "fmt"

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
