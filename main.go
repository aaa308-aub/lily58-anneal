package main

import (
	"fmt"
	"path"
	"unicode"

	"github.com/aaa308-aub/lily58-anneal/assets"
	"github.com/aaa308-aub/lily58-anneal/config"
)

func main() {

	keyConfig := config.KeyConfig
	langDataFileName := config.TargetLanguageCode + ".tsv" // The same for all N-grams.
	const numKeys, numSymbols = len(keyConfig), len(config.TargetSymbols)
	var targetSymbols = func() [numSymbols]rune {
		var ts [numSymbols]rune
		for i, symbol := range config.TargetSymbols {
			ts[i] = symbol
		}
		return ts
	}()

	if numKeys < 2 || numKeys > 29 {
		panic(fmt.Errorf(
			"number of included keys (%d) must be between 2 and 29 inclusive",
			numKeys,
		))
	}

	numKeysIncluded := 0
	for i := range keyConfig {
		if keyConfig[i].AssignedFinger != config.FingerNil {
			numKeysIncluded += 1
		}
	}

	if numSymbols != numKeysIncluded {
		panic(fmt.Errorf(
			"number of target symbols (%d) does not match number of included keys (%d)",
			numSymbols,
			numKeysIncluded,
		))
	}

	for i := range targetSymbols {
		targetSymbols[i] = unicode.ToLower(targetSymbols[i])
	}

	langDataFilePath := path.Join("assets", "counts", "monograms", langDataFileName)
	var monogramFreqs [numSymbols]float32
	err := assets.GetMonogramData(langDataFilePath, targetSymbols[:], monogramFreqs[:])
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse monogram data: %w",
			err,
		))
	}
	for i, freq := range monogramFreqs {
		if freq == 0 {
			panic(fmt.Errorf(
				"found symbol (%q) with zero-frequency in monogram data (%q)",
				targetSymbols[i],
				langDataFileName,
			))
		}
	}

	langDataFilePath = path.Join("assets", "counts", "bigrams", langDataFileName)
	var bigramFreqs [numSymbols * numSymbols]float32
	err = assets.GetBigramData(langDataFilePath, targetSymbols[:], bigramFreqs[:])
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse bigram data: %w",
			err,
		))
	}
}
