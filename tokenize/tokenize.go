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
const MULTI_THREAD = true

var begin time.Time

const REPORT = true

type FreqCollectorState struct {
	tokensMu       sync.Mutex
	tokensInCursor int
	tokensIn       []rune
}

func (state *FreqCollectorState) Collect() map[Pair]int {
	freq := map[Pair]int{}

	var begin, end int
	for {
		state.tokensMu.Lock()
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

func Report(label string, avgSum *time.Duration, iteration int) {
	if REPORT {
		since := time.Since(begin)
		*avgSum += since

		average := time.Duration(avgSum.Nanoseconds() / int64(iteration+1))

		fmt.Printf("\t%v: %v average %v\n", label, since, average)
		begin = time.Now()
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

	begin = time.Now()

	var workers = make(chan map[Pair]int)
	var iterationBegin time.Time

	var findingPairsSum time.Duration
	var bestPairsSum time.Duration
	var replacingSum time.Duration
	var iterationSum time.Duration

	for i := range 256 {
		pairs = append(pairs, Pair{
			L: rune(i),
		})
	}

	for _, r := range text {
		state.tokensIn = append(state.tokensIn, r)
	}

	tokenLenBegin := len(state.tokensIn)

	fmt.Printf("Tokenize\n")

	for range 250 {
		iterationBegin = time.Now()

		if REPORT {
			fmt.Printf("Iteration %d\n", iteration)
		}

		if MULTI_THREAD {
			for range FREQ_WORKER_COUNT {
				go func() {
					freq := state.Collect()
					workers <- freq
				}()
			}

			for range FREQ_WORKER_COUNT {
				workerFreq := <-workers
				for pair, count := range workerFreq {
					freq[pair] += count
				}
			}
		} else {
			freqCollect(0, len(state.tokensIn), &state.tokensIn, &freq)
		}

		Report("Finding pairs", &findingPairsSum, iteration)

		var best_pair Pair
		for pair, count := range freq {
			if count > freq[best_pair] {
				best_pair = pair
			}
		}

		Report("Finding best pair", &bestPairsSum, iteration)

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
		Report("Replacing Pairs", &replacingSum, iteration)
		if REPORT {
			since := time.Since(iterationBegin)
			iterationSum = iterationSum + since
			average := time.Duration(iterationSum.Nanoseconds() / int64(iteration+1))
			fmt.Printf("\tIteration time: %v average %v\n", time.Since(iterationBegin), average)
		}

		state.tokensIn = tokensOut
		state.tokensInCursor = 0

		tokensOut = nil
		iteration++
		clear(freq)
	}
	compression := int(float32(tokenLenBegin-len(state.tokensIn)) / float32(tokenLenBegin) * 100)
	fmt.Printf("Start char count: %d\n", tokenLenBegin)
	fmt.Printf("Iterations:       %v\n", iteration)
	fmt.Printf("End char count:   %d\n", len(state.tokensIn))
	fmt.Printf("Compression:      %v%%\n", compression)
	fmt.Printf("Pair count:       %d\n", len(pairs)-256)

	return pairs
}
