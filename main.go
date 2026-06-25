package main

import (
	"fmt"
	"math/rand/v2"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/aaa308-aub/lily58-anneal/annealing"
	"github.com/aaa308-aub/lily58-anneal/assets"
	msg "github.com/aaa308-aub/lily58-anneal/assets/messages"
	cfg "github.com/aaa308-aub/lily58-anneal/config"
)

var validFlagsMode = map[string]struct{}{
	"anneal": {}, "bruteforce": {},
}

func main() {

	const mode = cfg.Mode

	if _, ok := validFlagsMode[mode]; !ok {
		panic(fmt.Errorf("invalid mode set (%q)", mode))
	}

	const numSymbols = cfg.NumSymbols

	{ // Closure hides numKeysIncluded.
		numKeysIncluded := 0
		for _, key := range cfg.KeyInfos {
			if key.AssignedFinger != cfg.FingerNil {
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

	const langDataFileName = cfg.TargetLanguageCode + ".tsv" // The same for all N-grams.

	langDataFilePath := path.Join("assets", "counts", "monograms", langDataFileName)
	var monogramFreqs [numSymbols]float32
	err := assets.GetMonogramData(langDataFilePath, cfg.TargetSymbols[:], monogramFreqs[:])
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
				cfg.TargetSymbols[i],
			))
		}
	}

	langDataFilePath = path.Join("assets", "counts", "bigrams", langDataFileName)
	var bigramFreqs [numSymbols * numSymbols]float32
	err = assets.GetBigramData(langDataFilePath, cfg.TargetSymbols[:], bigramFreqs[:])
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse bigram data: %w",
			err,
		))
	}
	// Symmetrize matrix by aggregating (i,j) and (j,i).
	for i := range numSymbols {
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
	const numTopTrigrams = cfg.NumTopTrigrams
	var trigramInfos [numTopTrigrams]trigramInfo
	var symbolToTrigrams [numSymbols]uint64
	if !cfg.IgnoreTrigrams {
		err = assets.GetTrigramData(
			langDataFilePath,
			cfg.TargetSymbols[:],
			trigramInfos[:],
			numTopTrigrams,
		)
		if err != nil {
			panic(fmt.Errorf(
				"failed to parse trigram data: %w",
				err,
			))
		}

		assets.MapSymbolsToTrigrams(
			symbolToTrigrams[:],
			trigramInfos[:],
		)
	}

	// Filter out the excluded keys.
	const numKeys = cfg.NumKeys
	var keys [numSymbols]cfg.KeyInfo
	for i, j := 0, 0; i < numKeys; i++ {
		key := cfg.KeyInfos[i]
		if key.AssignedFinger != cfg.FingerNil {
			keys[j] = key
			j++
		}
	}

	layout := func() [numSymbols]int { // equals [0, 1, 2... numSymbols-1].
		var l [numSymbols]int
		for i := range l {
			l[i] = i
		}
		return l
	}()

	// Note: A layout could mean either a mapping of keys to symbols or a
	// mapping of symbols to keys (it's a bijection). So to be clear, if
	// layout[5] = 3, the 3rd key is mapped to the 5th symbol.

	switch mode {
	case "anneal":
		{

			var printerWG sync.WaitGroup
			var workerWG sync.WaitGroup
			printerWG.Add(1)
			go msg.PrintMessages(msg.MainChannel, &printerWG)

			seed := uint64(time.Now().UnixNano())
			for id := range runtime.NumCPU() {

				stream := uint64(id)
				source := rand.NewPCG(seed, stream)
				localRand := rand.New(source)

				localRand.Shuffle(numSymbols, func(i, j int) {
					layout[i], layout[j] = layout[j], layout[i]
				})

				annealingInputs := annealing.AnnealingInputs{
					Layout:           layout,
					KeyInfos:         keys,
					MonogramFreqs:    monogramFreqs,
					BigramFreqs:      bigramFreqs,
					TrigramInfos:     trigramInfos,
					SymbolToTrigrams: symbolToTrigrams,
				}

				workerWG.Add(1)
				go annealing.RunAnnealing(annealingInputs, id, localRand, &workerWG)
			}

			workerWG.Wait()
			close(msg.MainChannel)
			printerWG.Wait()
		}
	case "bruteforce": // W.I.P.
		{
		}
	}
}
