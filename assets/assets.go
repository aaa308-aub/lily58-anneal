package assets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Takes a monogram count data path, the target symbols, and a slice of
// monogram frequencies. Fills the slice for each symbol, matching
// by index.
func GetMonogramData(
	targetLanguagePath string,
	targetSymbols []rune,
	monogramFreqs []float32,
) error {

	file, err := os.Open(targetLanguagePath)
	if err != nil {
		return fmt.Errorf("failed to open language data file: %w", err)
	}
	if filepath.Ext(targetLanguagePath) != ".tsv" {
		return fmt.Errorf("target language data file must be a .tsv file")
	}
	defer file.Close()

	monogramCounts, totalCount := make([]int, len(targetSymbols)), 0
	scanner, lineNumber := bufio.NewScanner(file), 0
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()
		if line == "" {
			return fmt.Errorf(
				"found empty line (%d) in vetted data file outside of EOF",
				lineNumber,
			)
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return fmt.Errorf(
				"found line (%d) in vetted data file with %d parts, expected only 2",
				lineNumber,
				len(parts),
			)
		}

		runes := []rune(parts[0])
		if len(runes) != 1 {
			return fmt.Errorf(
				"found line (%d) in vetted data file with a %d-rune monogram",
				lineNumber,
				len(runes),
			)
		}
		monogram := runes[0]
		count, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf(
				"failed to parse monogram count via strconv.Atoi in line %d: %w",
				lineNumber,
				err,
			)
		}
		if count <= 0 {
			return fmt.Errorf(
				"found unexpected, non-positive monogram count (%d) in line %d",
				count,
				lineNumber,
			)
		}

		for i, symbol := range targetSymbols { // A simple linear search is fine here.
			if symbol == monogram {
				monogramCounts[i] = count
				totalCount += count
				break
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	if totalCount <= 0 {
		return fmt.Errorf(
			"total count of monograms is %d, expected a strictly positive value",
			totalCount,
		)
	}
	for i, count := range monogramCounts {
		monogramFreqs[i] = float32(count) / float32(totalCount)
	}

	return nil
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

	file, err := os.Open(targetLanguagePath)
	if err != nil {
		return fmt.Errorf("failed to open language data file: %w", err)
	}
	if filepath.Ext(targetLanguagePath) != ".tsv" {
		return fmt.Errorf("target language data file must be a .tsv file")
	}
	defer file.Close()

	numSymbols := len(targetSymbols)

	symbolToIndex := make(map[rune]int, numSymbols)
	for i, symbol := range targetSymbols {
		symbolToIndex[symbol] = i
	}

	bigramCounts, totalCount := make([]int, numSymbols*numSymbols), 0
	scanner, lineNumber := bufio.NewScanner(file), 0
	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()
		if line == "" {
			return fmt.Errorf(
				"found empty line (%d) in vetted data file outside of EOF",
				lineNumber,
			)
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return fmt.Errorf(
				"found line (%d) in vetted data file with %d parts, expected only 2",
				lineNumber,
				len(parts),
			)
		}

		runes := []rune(parts[0])
		if len(runes) != 2 {
			return fmt.Errorf(
				"found line (%d) in vetted data file with a %d-rune bigram",
				lineNumber,
				len(runes),
			)
		}

		count, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf(
				"failed to parse bigram count via strconv.Atoi in line %d: %w",
				lineNumber,
				err,
			)
		}
		if count <= 0 {
			return fmt.Errorf(
				"found unexpected, non-positive bigram count (%d) in line %d",
				count,
				lineNumber,
			)
		}
		index1, ok1 := symbolToIndex[runes[0]]
		index2, ok2 := symbolToIndex[runes[1]]
		// Decided to do this check *after* validating the count. Checking before
		// could lead to missing file corruption, all for a tiny performance gain.
		if !ok1 || !ok2 {
			continue
		}

		matrixIndex := numSymbols*index1 + index2
		bigramCounts[matrixIndex] = count
		totalCount += count
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	if totalCount <= 0 {
		return fmt.Errorf(
			"total count of bigrams is %d, expected a strictly positive value",
			totalCount,
		)
	}
	for i, count := range bigramCounts {
		bigramFreqs[i] = float32(count) / float32(totalCount)
	}

	return nil
}

// TODO: refactor and split this into two functions.
func GetTrigramData(
	targetLanguagePath string,
	targetSymbols []rune,
	trigramFreqs []float32,
	symbolToTrigramIndex []int8,
	numTopTrigrams int8,
) error {

	file, err := os.Open(targetLanguagePath)
	if err != nil {
		return fmt.Errorf("failed to open language data file: %w", err)
	}
	if filepath.Ext(targetLanguagePath) != ".tsv" {
		return fmt.Errorf("target language data file must be a .tsv file")
	}
	defer file.Close()

	numSymbols := len(targetSymbols)

	symbolToIndex := make(map[rune]int, numSymbols)
	for i, symbol := range targetSymbols {
		symbolToIndex[symbol] = i
	}

	// Use -1 as a flag for empty slots in a symbol's trigram indices.
	for i, _ := range symbolToTrigramIndex {
		symbolToTrigramIndex[i] = -1
	}

	trigramCounts, totalCount := make([]int, numTopTrigrams), 0
	countsIndex := int8(0) // lineNumber can't be used if some trigrams are skipped.
	scanner, lineNumber := bufio.NewScanner(file), 0
	for scanner.Scan() && countsIndex < numTopTrigrams {
		lineNumber++

		line := scanner.Text()
		if line == "" {
			return fmt.Errorf(
				"found empty line (%d) in vetted data file outside of EOF",
				lineNumber,
			)
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return fmt.Errorf(
				"found line (%d) in vetted data file with %d parts, expected only 2",
				lineNumber,
				len(parts),
			)
		}

		runes := []rune(parts[0])
		if len(runes) != 3 {
			return fmt.Errorf(
				"found line (%d) in vetted data file with a %d-rune trigram",
				lineNumber,
				len(runes),
			)
		}

		count, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf(
				"failed to parse trigram count via strconv.Atoi in line %d: %w",
				lineNumber,
				err,
			)
		}
		if count <= 0 {
			return fmt.Errorf(
				"found unexpected, non-positive trigram count (%d) in line %d",
				count,
				lineNumber,
			)
		}
		index1, ok1 := symbolToIndex[runes[0]]
		index2, ok2 := symbolToIndex[runes[1]]
		index3, ok3 := symbolToIndex[runes[2]]
		// Decided to do this check *after* validating the count. Checking before
		// could lead to missing file corruption, all for a tiny performance gain.
		if !ok1 || !ok2 || !ok3 {
			continue
		}

		trigramCounts[countsIndex] = count
		totalCount += count

		indices := []int{index1, index2, index3}
		for _, index := range indices {
			offset := index * numSymbols
			for i := offset; i < offset+int(numTopTrigrams); i++ {
				if symbolToTrigramIndex[i] == -1 {
					symbolToTrigramIndex[i] = countsIndex
					break
				}
			}
		}
		countsIndex++
	}

	// Notice: the loop could theoretically and silently end before
	// countsIndex reaches numTopTrigrams. It won't break the
	// logic, just that the number of trigrams recorded is less
	// than requested.

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	if totalCount <= 0 {
		return fmt.Errorf(
			"total count of trigrams is %d, expected a strictly positive value",
			totalCount,
		)
	}
	for i, count := range trigramCounts {
		trigramFreqs[i] = float32(count) / float32(totalCount)
	}

	return nil
}
