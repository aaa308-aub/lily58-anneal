package main

import (
	"fmt"
	"math/rand/v2"
	"path"
	"runtime"

	//"sync"
	"time"
	"unicode"

	"github.com/aaa308-aub/lily58-anneal/annealing"
	"github.com/aaa308-aub/lily58-anneal/assets"
	"github.com/aaa308-aub/lily58-anneal/config"
)

func main() {
	const numSymbols = len(config.TargetSymbols)

	{ // Closure hides numKeysIncluded.
		numKeysIncluded := 0
		for _, key := range config.KeyInfos {
			if key.AssignedFinger != config.FingerNil {
				numKeysIncluded++
			}
		}

		if numKeysIncluded < 3 || numKeysIncluded > 29 {
			panic(fmt.Errorf(
				"number of included keys (%d) must be between 3 and 29 inclusive",
				numKeysIncluded,
			))
		}

		if numKeysIncluded != numSymbols {
			panic(fmt.Errorf(
				"number of included keys (%d) is different from number of symbols (%d)",
				numKeysIncluded,
				numSymbols,
			))
		}
	}

	var targetSymbols = func() [numSymbols]rune {
		var ts [numSymbols]rune
		for i, symbol := range config.TargetSymbols {
			ts[i] = unicode.ToLower(symbol)
		}
		return ts
	}()

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
	// Symmetrize matrix by aggregating (i,j) and (j,i).
	for i := 0; i < numSymbols; i++ {
		for j := i + 1; j < numSymbols; j++ {
			entryIndex := numSymbols*i + j
			transposedIndex := numSymbols*j + i
			aggregate := bigramFreqs[entryIndex] + bigramFreqs[transposedIndex]

			// Assign to both sides
			bigramFreqs[entryIndex] = aggregate
			bigramFreqs[transposedIndex] = aggregate
		}
	}

	type trigramInfo = assets.TrigramInfo
	langDataFilePath = path.Join("assets", "counts", "trigrams", langDataFileName)
	const numTopTrigrams = config.NumTopTrigrams
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

	var symbolToTrigrams [numSymbols]uint64
	assets.MapSymbolsToTrigrams(
		symbolToTrigrams[:],
		trigramInfos[:],
	)

	// Filter out the excluded keys.
	const numKeys = config.NumKeys
	var keys [numSymbols]config.KeyInfo
	for i, j := 0, 0; i < numKeys; i++ {
		key := config.KeyInfos[i]
		if key.AssignedFinger != config.FingerNil {
			keys[j] = key
			j++
		}
	}

	var identityLayout [numSymbols]int // equals [0, 1, 2... numSymbols-1]
	for i := range identityLayout {
		identityLayout[i] = i
	}

	// Note: A layout could mean either a mapping of keys to symbols or a
	// mapping of symbols to keys (it's a bijection). So to be clear, if
	// layout[5] = 3, the 3rd key is mapped to the 5th symbol.

	//var wg sync.WaitGroup
	seed := uint64(time.Now().UnixNano())
	for i := 0; i < runtime.NumCPU(); i++ {
		//wg.Add(1)

		stream := uint64(i)
		source := rand.NewPCG(seed, stream)
		localRand := rand.New(source)

		layout := identityLayout
		localRand.Shuffle(numSymbols, func(i, j int) {
			layout[i], layout[j] = layout[j], layout[i]
		})

		params := annealing.CoupleAnnealingParams(
			&layout,
			&keys,
			&monogramFreqs,
			&bigramFreqs,
			&trigramInfos,
			&symbolToTrigrams,
		)

		cost := annealing.InitialLayoutCost(params)
		fmt.Println(layout, cost)
	}
	//wg.Wait()
}
