package assets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Note: these functions could be joined into one "GetNgramData" func, but it's
// probably better to split them.

// Takes as input the target language .tsv data path, the target symbols, and
// a slice of monogram frequencies. Fills the slice for each symbol by index.
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

	monogramCounts := make([]int, len(targetSymbols))
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
				"found line (%d) in vetted data file with %d runes, expected only 1",
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
				break
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("non-EOF error encountered by scanner: %w", err)
	}

	// Some redundancy here, but it won't matter for such tiny slices.
	totalCount := 0
	for _, count := range monogramCounts {
		totalCount += count
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
