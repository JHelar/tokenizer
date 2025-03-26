package tokenize

import (
	"sync"
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
	freqMu         sync.Mutex
	freq           map[Pair]int
}

func (state *FreqCollectorState) Collect() {
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

		if begin >= end {
			break
		}

		freqCollect(begin, end, &state.tokensIn, &freq)
	}

	state.freqMu.Lock()
	for pair, value := range freq {
		state.freq[pair] += value
	}
	state.freqMu.Unlock()
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
		freq:           make(map[Pair]int),
	}
	iteration := 0

	var wg sync.WaitGroup

	for i := range 256 {
		pairs = append(pairs, Pair{
			L: rune(i),
		})
	}

	for _, r := range text {
		state.tokensIn = append(state.tokensIn, r)
	}

	for {
		// fmt.Printf("Iteration %d\n", iteration)
		// fmt.Printf("\tToken count: %d\n", len(state.tokensIn))
		// begin := time.Now()

		for range FREQ_WORKER_COUNT {
			wg.Add(1)
			go func() {
				defer wg.Done()
				state.Collect()
			}()
		}
		wg.Wait()

		// fmt.Printf("\tFinding pairs: %v\n", time.Since(begin))
		// begin = time.Now()

		var best_pair Pair
		for pair, count := range state.freq {
			if count > state.freq[best_pair] {
				best_pair = pair
			}
		}

		// fmt.Printf("\tFinding best pair: %v\n", time.Since(begin))
		// begin = time.Now()

		if state.freq[best_pair] <= 1 {
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
		clear(state.freq)
	}

	return pairs
}
