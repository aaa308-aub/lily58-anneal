package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
)

// To keep the matrix an array instead of a slice, its size must be const
// I'll just fill the largest possible matrix, memory waste is negligible
const maxNumberOfKeys = 29
const maxBigramMatrixSize = maxNumberOfKeys * maxNumberOfKeys
const minFileSize = 1024 // bytes
type bigramMatrix = [maxBigramMatrixSize]int

func FillBigramMatrix(filepath string, targetSymbols []rune) (bigramMatrix, error) {

	var matrix bigramMatrix
	numberOfSymbols := len(targetSymbols)

	if numberOfSymbols > maxNumberOfKeys {
		err := fmt.Errorf(
			"Error: FillBigramMatrix called with number of symbols (%d) greater than max possible number of keys (%d)",
			numberOfSymbols,
			maxNumberOfKeys,
		)

		return bigramMatrix{}, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return bigramMatrix{}, err
	}
	defer file.Close()

	fileSize, err := os.Stat(filepath)
	if err != nil {
		return bigramMatrix{}, err
	}
	fileSize_int := int(fileSize.Size())
	if fileSize_int < minFileSize {
		err := fmt.Errorf(
			"Error: Data file size must be at least 1KB. Received file of size %d bytes",
			fileSize_int,
		)

		return bigramMatrix{}, err
	}

	symbolToIndex := make(map[rune]int, numberOfSymbols)
	for i := range targetSymbols {
		symbol := unicode.ToLower(targetSymbols[i])
		symbolToIndex[symbol] = i
	}

	reader := bufio.NewReader(file)
	prev, isPrev := rune(0), false
	for {
		curr, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}

			return bigramMatrix{}, err
		}
		curr = unicode.ToLower(curr)

		currIndex, ok := symbolToIndex[curr]
		if !ok {
			isPrev = false
			continue
		}

		if isPrev {
			prevIndex := symbolToIndex[prev]
			matrix[prevIndex*numberOfSymbols+currIndex] += 1
		}

		prev, isPrev = curr, true
	}

	return matrix, nil
}
