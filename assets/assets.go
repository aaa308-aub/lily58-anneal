package assets

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// lineIdx is passed for specifying where an error occurs. Make sure to pass
// lineIdx as one-indexed to match human debugging.
func parseNgramLine(
	line string,
	lineIdx int,
	nGramSize int,
) ([]rune, int, error) {

	parts := strings.Split(line, "\t")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf(
			"found line (%d) in vetted data file with %d parts, expected "+
				"strictly 2",
			lineIdx,
			len(parts),
		)
	}

	runes := []rune(parts[0])
	if len(runes) != nGramSize {
		return nil, 0, fmt.Errorf(
			"found line (%d) in vetted data file with a %d-rune n-gram, "+
				"expected %d",
			lineIdx,
			len(runes),
			nGramSize,
		)
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf(
			"failed to parse n-gram count in line %d: %w",
			lineIdx,
			err,
		)
	}
	if count <= 0 {
		return nil, 0, fmt.Errorf(
			"found unexpected, non-positive n-gram count (%d) in line %d",
			count,
			lineIdx,
		)
	}

	return runes, count, nil
}

func turnCountsToFreqs(counts, freqs []float32) error {

	total := float32(0)
	for _, c := range counts {
		total += c
	}

	if total < 0 {
		return fmt.Errorf(
			"total count of n-grams is negative, overflow possible but " +
				"unlikely cause",
		)
	}

	if total == 0 {
		return fmt.Errorf(
			"total count of n-grams is zero, data file may be empty",
		)
	}

	for i, c := range counts {
		freqs[i] = c / total
	}

	return nil
}

func readFileAsLines(path string) ([]string, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read language data file: %w", err)
	}

	// Remove any trailing empty lines. Empty lines in-between others are
	// treated as faulty.
	linesConcat := strings.TrimSpace(string(data))
	lines := strings.Split(linesConcat, "\n")
	return lines, nil
}

// Frequencies match symbols by index.
func GetMonogramData(
	path string,
	syms []rune,
	freqs []float32,
) error {

	lines, err := readFileAsLines(path)
	if err != nil {
		return err
	}
	counts := make([]float32, len(freqs))
	for lineIdx, line := range lines {

		runes, count, err := parseNgramLine(line, lineIdx+1, 1)
		if err != nil {
			return err
		}

		monogram := runes[0]
		for i, symbol := range syms { // A simple linear search is fine here.
			if symbol == monogram {
				counts[i] = float32(count)
				break
			}
		}
	}

	err = turnCountsToFreqs(counts, freqs)
	return err
}

// Slice of frequencies should view a flattened 2D LUT that
// cross-tabulates slice of symbols with itself.
func GetBigramData(
	path string,
	syms []rune,
	freqs []float32,
) error {

	nSym := len(syms)
	symToIdx := make(map[rune]int, nSym)
	for i, sym := range syms {
		symToIdx[sym] = i
	}

	lines, err := readFileAsLines(path)
	if err != nil {
		return err
	}
	counts := make([]float32, len(freqs))
	for lineIdx, line := range lines {

		runes, count, err := parseNgramLine(line, lineIdx+1, 2)
		if err != nil {
			return err
		}

		i, ok1 := symToIdx[runes[0]]
		j, ok2 := symToIdx[runes[1]]
		if !ok1 || !ok2 || i == j { // Bigram symbols must be distinct.
			continue
		}

		idx := nSym*i + j
		counts[idx] = float32(count)
	}

	err = turnCountsToFreqs(counts, freqs)
	return err
}

// Syms field contains the symbols by index and in order within trigram.
type TrigramT struct {
	Freq float32
	Syms [3]int8
}

// nTrigram is always passed as 64 for bitmasking optimization unless modified.
func GetTrigramData(
	path string,
	syms []rune,
	trigrams []TrigramT,
	nTrigram int8,
) error {

	nSym := len(syms)
	symToIdx := make(map[rune]int8, nSym)
	for i, symbol := range syms {
		symToIdx[symbol] = int8(i)
	}

	lines, err := readFileAsLines(path)
	if err != nil {
		return err
	}
	trigramSyms := make([][3]int8, nTrigram)
	counts := make([]float32, nTrigram)
	countsIdx := int8(0) // lineIdx can't be used if some trigrams are skipped.
	for lineIdx, line := range lines {
		if countsIdx >= nTrigram {
			break
		}

		runes, count, err := parseNgramLine(line, lineIdx+1, 3)
		if err != nil {
			return err
		}

		i, ok1 := symToIdx[runes[0]]
		j, ok2 := symToIdx[runes[1]]
		k, ok3 := symToIdx[runes[2]]
		if !ok1 || !ok2 || !ok3 {
			continue
		}
		// A trigram with non-distinct symbols is invalid and should be ignored.
		if i == j || i == k || j == k {
			continue
		}

		trigramSyms[countsIdx] = [3]int8{i, j, k}
		counts[countsIdx] = float32(count)
		countsIdx++
	}

	// Notice: the loop could theoretically and silently end before
	// countsIdx reaches nTrigrams. It won't break the logic, just
	// that the number of trigrams recorded is less than requested.

	freqs := make([]float32, len(counts))
	err = turnCountsToFreqs(counts, freqs)
	if err != nil {
		return err
	}

	for i := range nTrigram {
		trigrams[i] = TrigramT{freqs[i], trigramSyms[i]}
	}

	return nil
}

// Takes trigramInfos from GetTrigramData to map each symbol using a bitmask
// to the indices of trigrams it belongs to.
func MapSymbolsToTrigrams(
	symToTrigs []uint64,
	trigrams []TrigramT,
) {

	for i, t := range trigrams {
		bit := uint64(1 << i)

		if t == (TrigramT{}) {
			break
		}

		for _, sym := range t.Syms {
			symToTrigs[sym] |= bit
		}
	}
}
