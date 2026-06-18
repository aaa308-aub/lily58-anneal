package assets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func openLanguageDataFile(targetLanguagePath string) (*os.File, error) {

	file, err := os.Open(targetLanguagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open language data file: %w", err)
	}
	if filepath.Ext(targetLanguagePath) != ".tsv" {
		return nil, fmt.Errorf("target language data file must be a .tsv file")
	}
	return file, nil
}

func parseNgramLine(
	line string,
	lineNumber int,
	nGramSize int,
) ([]rune, int, error) {

	if line == "" {
		return nil, 0, fmt.Errorf(
			"found empty line (%d) in vetted data file outside of EOF",
			lineNumber,
		)
	}

	parts := strings.Split(line, "\t")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf(
			"found line (%d) in vetted data file with %d parts, expected only 2",
			lineNumber,
			len(parts),
		)
	}

	runes := []rune(parts[0])
	if len(runes) != nGramSize {
		return nil, 0, fmt.Errorf(
			"found line (%d) in vetted data file with a %d-rune n-gram, expected %d",
			lineNumber,
			len(runes),
			nGramSize,
		)
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf(
			"failed to parse n-gram count via strconv.Atoi in line %d: %w",
			lineNumber,
			err,
		)
	}
	if count <= 0 {
		return nil, 0, fmt.Errorf(
			"found unexpected, non-positive n-gram count (%d) in line %d",
			count,
			lineNumber,
		)
	}

	return runes, count, nil
}

func turnCountsToFreqs(counts []float32) error {

	totalCount := float32(0)
	for _, count := range counts {
		totalCount += count
	}

	if totalCount < 0 {
		return fmt.Errorf(
			"total count of n-grams is negative, overflow possible but unlikely cause",
		)
	}

	if totalCount == 0 {
		return fmt.Errorf(
			"total count of n-grams is zero, data file may be empty",
		)
	}

	for i, count := range counts {
		counts[i] = count / totalCount
	}
	return nil
}

// Takes a monogram count data path, the target symbols, and a slice of
// monogram frequencies. Fills the slice for each symbol, matching
// by index.
func GetMonogramData(
	targetLanguagePath string,
	targetSymbols []rune,
	monogramFreqs []float32,
) error {

	file, err := openLanguageDataFile(targetLanguagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner, lineNumber := bufio.NewScanner(file), 0
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()
		runes, count, err := parseNgramLine(line, lineNumber, 1)
		if err != nil {
			return err
		}

		monogram := runes[0]
		for i, symbol := range targetSymbols { // A simple linear search is fine here.
			if symbol == monogram {
				monogramFreqs[i] = float32(count)
				break
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	err = turnCountsToFreqs(monogramFreqs)
	return err
}

// Takes a bigram count data path, the target symbols, and a slice of
// bigram frequencies to fill.
//
// This slice is treated as a flattened 2D matrix by cross-tabulating
// the target symbols with itself for fast bigram frequency lookup.
func GetBigramData(
	targetLanguagePath string,
	targetSymbols []rune,
	bigramFreqs []float32,
) error {

	file, err := openLanguageDataFile(targetLanguagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	numSymbols := len(targetSymbols)

	symbolToIndex := make(map[rune]int, numSymbols)
	for i, symbol := range targetSymbols {
		symbolToIndex[symbol] = i
	}

	scanner, lineNumber := bufio.NewScanner(file), 0
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()
		runes, count, err := parseNgramLine(line, lineNumber, 2)
		if err != nil {
			return err
		}

		index1, ok1 := symbolToIndex[runes[0]]
		index2, ok2 := symbolToIndex[runes[1]]
		if !ok1 || !ok2 {
			continue
		}

		matrixIndex := numSymbols*index1 + index2
		bigramFreqs[matrixIndex] = float32(count)
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	err = turnCountsToFreqs(bigramFreqs)
	return err
}

// The ordering is required by the engine. There's no way around this.
type TrigramInfo struct {
	Freq                 float32
	orderedSymbolIndices [3]int8
}

// Takes a trigram count data path, the target symbols, and a slice of
// trigramInfos to fill with the top X trigrams, where X = numTopTrigrams.
func GetTrigramData(
	targetLanguagePath string,
	targetSymbols []rune,
	trigramInfos []TrigramInfo,
	numTopTrigrams int8,
) error {

	file, err := openLanguageDataFile(targetLanguagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	numSymbols := len(targetSymbols)

	symbolToIndex := make(map[rune]int, numSymbols)
	for i, symbol := range targetSymbols {
		symbolToIndex[symbol] = i
	}

	trigrams := make([][3]int8, numTopTrigrams)
	counts := make([]float32, numTopTrigrams)
	countsIndex := int8(0) // lineNumber can't be used if some trigrams are skipped.
	scanner, lineNumber := bufio.NewScanner(file), 0
	for scanner.Scan() && countsIndex < numTopTrigrams {
		lineNumber++

		line := scanner.Text()
		runes, count, err := parseNgramLine(line, lineNumber, 3)
		if err != nil {
			return err
		}

		index1, ok1 := symbolToIndex[runes[0]]
		index2, ok2 := symbolToIndex[runes[1]]
		index3, ok3 := symbolToIndex[runes[2]]
		if !ok1 || !ok2 || !ok3 {
			continue
		}

		trigrams[countsIndex] = [3]int8{int8(index1), int8(index2), int8(index3)}
		counts[countsIndex] = float32(count)
		countsIndex++
	}

	// Notice: the loop could theoretically and silently end before
	// countsIndex reaches numTopTrigrams. It won't break the
	// logic, just that the number of trigrams recorded is less
	// than requested.

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	err = turnCountsToFreqs(counts)
	if err != nil {
		return err
	}

	for i := int8(0); i < numTopTrigrams; i++ {
		trigramInfos[i] = TrigramInfo{counts[i], trigrams[i]}
	}

	return nil
}

// Takes trigramInfos from GetTrigramData to map each symbol to a bucket
// of indices of trigrams it belongs to in a flattened but sparse matrix.
//
// The value -1 is used as a flag for empty slots.
func MapSymbolsToTrigrams(
	symbolToTrigramIndex []int8,
	trigramInfos []TrigramInfo,
	numSymbols, numTrigrams int,
) error {

	if len(symbolToTrigramIndex) != numSymbols*numTrigrams {
		return fmt.Errorf(
			"length of symbolToTrigramIndex is %d, expected %d",
			len(symbolToTrigramIndex),
			numSymbols*numTrigrams,
		)
	}

	if len(trigramInfos) != numTrigrams {
		return fmt.Errorf(
			"length of trigramFreqs (%d) does not match number of numTrigrams (%d)",
			len(trigramInfos),
			numTrigrams,
		)
	}

	for i := range symbolToTrigramIndex {
		symbolToTrigramIndex[i] = -1
	}

	for i, trigram := range trigramInfos {
		for _, symbolIndex := range trigram.orderedSymbolIndices {
			offset := numSymbols * int(symbolIndex)
			for j := offset; j < offset+numTrigrams; j++ {
				if symbolToTrigramIndex[j] == -1 {
					symbolToTrigramIndex[j] = int8(i)
					break
				}
			}
		}
	}

	return nil
}
