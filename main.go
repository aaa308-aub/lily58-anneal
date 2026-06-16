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
	targetSymbols := config.TargetSymbols
	langDataFile := config.TargetLanguageCode + ".tsv" // The same for all N-grams.

	const numKeys, numSymbols = len(keyConfig), len(targetSymbols)

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

	langDataPath := path.Join("assets", "counts", "monograms", langDataFile)
	var monogramFreq [numSymbols]float32
	err := assets.GetMonogramData(langDataPath, targetSymbols[:], monogramFreq[:])
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse monogram data: %w",
			err,
		))
	}
	for i, freq := range monogramFreq {
		if freq == 0 {
			panic(fmt.Errorf(
				"found symbol (%q) with zero-frequency in monogram data (%q)",
				targetSymbols[i],
				langDataFile,
			))
		}
	}
}
