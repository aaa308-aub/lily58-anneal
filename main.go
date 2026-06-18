package main

import (
	"fmt"
	"path"
	"unicode"

	"github.com/aaa308-aub/lily58-anneal/assets"
	"github.com/aaa308-aub/lily58-anneal/config"
)

func main() {
	const numKeys, numSymbols = len(config.KeyConfig), len(config.TargetSymbols)
	const numKeysIncluded = config.NumKeysIncluded

	var targetSymbols = func() [numSymbols]rune {
		var ts [numSymbols]rune
		for i, symbol := range config.TargetSymbols {
			ts[i] = symbol
		}
		return ts
	}()

	{ // Closure hides numKeysIncludedActual.
		numKeysIncludedActual := 0
		for _, key := range config.KeyConfig {
			if key.AssignedFinger != config.FingerNil {
				numKeysIncludedActual++
			}
		}

		if numKeysIncluded != numKeysIncludedActual {
			panic(fmt.Errorf(
				"number of keys included (%d) miscounted as %d, config must be fixed",
				numKeysIncludedActual,
				numKeysIncluded,
			))
		}
	}

	if numKeys < 2 || numKeys > 29 {
		panic(fmt.Errorf(
			"number of included keys (%d) must be between 2 and 29 inclusive",
			numKeys,
		))
	}

	for i := range targetSymbols {
		targetSymbols[i] = unicode.ToLower(targetSymbols[i])
	}

	langDataFileName := config.TargetLanguageCode + ".tsv" // The same for all N-grams.

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
				"found symbol (%q) with no monogram data, may not belong to language",
				targetSymbols[i],
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

	type trigramInfo = assets.TrigramInfo
	langDataFilePath = path.Join("assets", "counts", "trigrams", langDataFileName)
	const numTopTrigrams = 100
	var trigramInfos [numTopTrigrams]trigramInfo
	err = assets.GetTrigramData(
		langDataFilePath,
		targetSymbols[:],
		trigramInfos[:],
		numTopTrigrams,
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse trigram data: %w",
			err,
		))
	}

	var symbolToTrigramIndex [numSymbols * numTopTrigrams]int8
	err = assets.MapSymbolsToTrigrams(
		symbolToTrigramIndex[:],
		trigramInfos[:],
		numSymbols,
		numTopTrigrams,
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to map symbols to trigrams: %w",
			err,
		))
	}
}
