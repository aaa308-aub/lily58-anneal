package assets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func openDataFile(filePath string) (*os.File, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open language data file: %w", err)
	}
	if filepath.Ext(filePath) != ".tsv" {
		return nil, fmt.Errorf("target language data file must be a .tsv file")
	}
	return file, nil
}

func parseNgramLine(
	line string,
	lineIdx int,
	nGramSize int,
) ([]rune, int, error) {

	if line == "" {
		return nil, 0, fmt.Errorf(
			"found empty line (%d) in vetted data file outside of EOF",
			lineIdx,
		)
	}

	parts := strings.Split(line, "\t")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf(
			"found line (%d) in vetted data file with %d parts, expected only 2",
			lineIdx,
			len(parts),
		)
	}

	runes := []rune(parts[0])
	if len(runes) != nGramSize {
		return nil, 0, fmt.Errorf(
			"found line (%d) in vetted data file with a %d-rune n-gram, expected %d",
			lineIdx,
			len(runes),
			nGramSize,
		)
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf(
			"failed to parse n-gram count via strconv.Atoi in line %d: %w",
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
			"total count of n-grams is negative, overflow possible but unlikely cause",
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

// Frequencies match symbols by index.
func GetMonogramData(
	filePath string,
	syms []rune,
	freqs []float32,
) error {

	file, err := openDataFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner, lineIdx := bufio.NewScanner(file), 0
	counts := make([]float32, len(freqs))
	for scanner.Scan() {
		lineIdx++

		line := scanner.Text()
		runes, count, err := parseNgramLine(line, lineIdx, 1)
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

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	err = turnCountsToFreqs(counts, freqs)
	return err
}

// Slice of frequencies should view a flattened 2D LUT that
// cross-tabulates slice of symbols with itself.
func GetBigramData(
	filePath string,
	syms []rune,
	freqs []float32,
) error {

	file, err := openDataFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	nSym := len(syms)

	symToIdx := make(map[rune]int, nSym)
	for i, sym := range syms {
		symToIdx[sym] = i
	}

	scanner, lineIdx := bufio.NewScanner(file), 0
	counts := make([]float32, len(freqs))
	for scanner.Scan() {
		lineIdx++

		line := scanner.Text()
		runes, count, err := parseNgramLine(line, lineIdx, 2)
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

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	err = turnCountsToFreqs(counts, freqs)
	return err
}

// Syms field should contain the symbols (by index) in their
// order within trigram.
type TrigramT struct {
	Freq float32
	Syms [3]int8
}

func GetTrigramData(
	filePath string,
	syms []rune,
	trigrams []TrigramT,
	nTrigram int8,
) error {

	file, err := openDataFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	nSym := len(syms)

	symToIdx := make(map[rune]int8, nSym)
	for i, symbol := range syms {
		symToIdx[symbol] = int8(i)
	}

	trigramSyms := make([][3]int8, nTrigram)
	counts := make([]float32, nTrigram)
	countsIdx := int8(0) // lineIdx can't be used if some trigrams are skipped.
	scanner, lineIdx := bufio.NewScanner(file), 0
	for scanner.Scan() && countsIdx < nTrigram {
		lineIdx++

		line := scanner.Text()
		runes, count, err := parseNgramLine(line, lineIdx, 3)
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

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

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
