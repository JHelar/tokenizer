package tokenize

import (
	"fmt"
	"sync"
	"time"
)

type Pair struct {
	L rune
	R rune
}

func (pair Pair) Equals(another Pair) bool {
	return pair.L == another.L && pair.R == another.R
}

const FREQ_COLLECTION_CHUNK_SIZE = 64 * 1024
const FREQ_WORKER_COUNT = 10

type FreqCollectorState struct {
	tokensMu       sync.Mutex
	tokensInCursor int
	tokensIn       []rune
}

func (state *FreqCollectorState) Collect() map[Pair]int {
	freq := make(map[Pair]int)

	for {
		state.tokensMu.Lock()
		var begin, end int
		begin = state.tokensInCursor

		if state.tokensInCursor+FREQ_COLLECTION_CHUNK_SIZE <= len(state.tokensIn) {
			state.tokensInCursor += FREQ_COLLECTION_CHUNK_SIZE
		} else {
			state.tokensInCursor = len(state.tokensIn)
		}

		end = state.tokensInCursor
		state.tokensMu.Unlock()

		if end <= begin {
			break
		}

		freqCollect(begin, end, &state.tokensIn, &freq)
	}

	return freq
}

func freqCollect(begin int, end int, tokensIn *[]rune, freq *map[Pair]int) {
	for i := begin; i < end; i++ {
		if i+1 >= len(*tokensIn) {
			break
		}
		key := Pair{
			L: (*tokensIn)[i],
			R: (*tokensIn)[i+1],
		}
		(*freq)[key]++
	}
}

func Tokenize(text string) []Pair {
	tokensOut := []rune{}
	pairs := []Pair{}
	state := FreqCollectorState{
		tokensIn:       []rune{},
		tokensInCursor: 0,
	}
	iteration := 0
	freq := make(map[Pair]int)

	var sem = make(chan map[Pair]int, FREQ_WORKER_COUNT)

	for i := range 256 {
		pairs = append(pairs, Pair{
			L: rune(i),
		})
	}

	for _, r := range text {
		state.tokensIn = append(state.tokensIn, r)
	}

	begin := time.Now()
	tokenLenBegin := len(state.tokensIn)
	fmt.Printf("Tokenize\n")
	fmt.Printf("\tStart char count: %d\n", len(state.tokensIn))

	for {
		for range FREQ_WORKER_COUNT {
			go func() {
				freq := state.Collect()
				sem <- freq
			}()
		}

		for range FREQ_WORKER_COUNT {
			workerFreq := <-sem
			for pair, count := range workerFreq {
				freq[pair] += count
			}
		}

		// freqCollect(0, len(state.tokensIn), &state.tokensIn, &mergedFreq)

		// fmt.Printf("\tFinding pairs: %v\n", time.Since(begin))
		// begin = time.Now()

		var best_pair Pair
		for pair, count := range freq {
			if count > freq[best_pair] {
				best_pair = pair
			}
		}

		// fmt.Printf("\tFinding best pair: %v\n", time.Since(begin))
		// begin = time.Now()

		if freq[best_pair] <= 1 {
			break
		}

		pairs = append(pairs, best_pair)

		for i := 0; i < len(state.tokensIn); i++ {
			if i+1 >= len(state.tokensIn) {
				tokensOut = append(tokensOut, state.tokensIn[i])
				break
			}

			pair := Pair{
				L: state.tokensIn[i],
				R: state.tokensIn[i+1],
			}

			if pair.Equals(best_pair) {
				tokensOut = append(tokensOut, rune(len(pairs)-1))
				i++
			} else {
				tokensOut = append(tokensOut, state.tokensIn[i])
			}
		}
		// fmt.Printf("\tReplacing Pairs: %v\n", time.Since(begin))

		state.tokensIn = tokensOut
		state.tokensInCursor = 0

		tokensOut = nil
		iteration++
		clear(freq)
	}
	compression := int(float32(tokenLenBegin-len(state.tokensIn)) / float32(tokenLenBegin) * 100)
	fmt.Printf("\tCompleted:        %v\n", time.Since(begin))
	fmt.Printf("\tIterations:       %v\n", iteration)
	fmt.Printf("\tEnd char count:   %d\n", len(state.tokensIn))
	fmt.Printf("\tCompression:      %v%%\n", compression)
	fmt.Printf("\tPair count:       %d\n", len(pairs)-256)

	return pairs
}
