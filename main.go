package main

import (
	"fmt"
	"math/rand/v2"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/aaa308-aub/lily58-anneal/anneal"
	"github.com/aaa308-aub/lily58-anneal/assets"
	msg "github.com/aaa308-aub/lily58-anneal/assets/messages"
	cfg "github.com/aaa308-aub/lily58-anneal/config"
)

var modes = map[string]struct{}{
	"anneal": {}, "bruteforce": {},
}

func main() {

	const mode = cfg.Mode

	if _, ok := modes[mode]; !ok {
		panic(fmt.Errorf("invalid mode set (%q)", mode))
	}

	const nSym = cfg.NumSymbols

	{ // Closure hides nKeysTargetted.
		nKeysTargetted := 0
		for _, key := range cfg.KeysAll {
			if key.Fin != cfg.FingerNil {
				nKeysTargetted++
			}
		}

		if nKeysTargetted < 3 || nKeysTargetted > 29 {
			panic(fmt.Errorf(
				"number of included keys (%d) must be between 3 and 29 inclusive",
				nKeysTargetted,
			))
		}

		if nKeysTargetted != nSym {
			panic(fmt.Errorf(
				"number of included keys (%d) is different from number of symbols (%d)",
				nKeysTargetted,
				nSym,
			))
		}
	}

	const fileName = cfg.TargetLanguageCode + ".tsv" // The same for all N-grams.
	var syms = cfg.SymbolsArr

	filePath := path.Join("assets", "counts", "monograms", fileName)
	var monoFreqs [nSym]float32
	err := assets.GetMonogramData(filePath, syms[:], monoFreqs[:])
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse monogram data: %w",
			err,
		))
	}
	for i, freq := range monoFreqs {
		if freq == 0 {
			panic(fmt.Errorf(
				"found symbol (%q) with no monogram data, may not belong to language",
				syms[i],
			))
		}
	}

	var biFreqs [nSym][nSym]float32
	{ // Closure hides biFreqsFlat.
		filePath = path.Join("assets", "counts", "bigrams", fileName)
		// Matrix biFreqs is flattened at first for easier coupling in assets.go.
		var biFreqsFlat [nSym * nSym]float32
		err = assets.GetBigramData(filePath, syms[:], biFreqsFlat[:])
		if err != nil {
			panic(fmt.Errorf(
				"failed to parse bigram data: %w",
				err,
			))
		}

		// Unflatten matrix and symmetrize it by aggregating (i,j) and (j,i).
		for i := range nSym {
			for j := i + 1; j < nSym; j++ {
				idx := nSym*i + j
				idxTrans := nSym*j + i
				aggregate := biFreqsFlat[idx] + biFreqsFlat[idxTrans]

				// Assign to both sides
				biFreqs[i][j] = aggregate
				biFreqs[j][i] = aggregate
			}
		}
	}

	type trigram = assets.TrigramT
	filePath = path.Join("assets", "counts", "trigrams", fileName)
	const nTrigrams = cfg.NumTopTrigrams
	var trigrams [nTrigrams]trigram
	var symToTrigs [nSym]uint64
	if !cfg.IgnoreTrigrams {
		err = assets.GetTrigramData(
			filePath,
			syms[:],
			trigrams[:],
			nTrigrams,
		)
		if err != nil {
			panic(fmt.Errorf(
				"failed to parse trigram data: %w",
				err,
			))
		}

		assets.MapSymbolsToTrigrams(
			symToTrigs[:],
			trigrams[:],
		)
	}

	// Filter out the excluded keys.
	const nKeyAll = cfg.NumKeysAll
	var keys [nSym]cfg.KeyT
	for i, j := 0, 0; i < nKeyAll; i++ {
		key := cfg.KeysAll[i]
		if key.Fin != cfg.FingerNil {
			keys[j] = key
			j++
		}
	}

	var layout [nSym]int // equals [0, 1, 2... numSymbols-1].
	for i := range layout {
		layout[i] = i
	}

	// Note: A layout could mean either a mapping of keys to symbols or a
	// mapping of symbols to keys (it's a bijection). So to be clear, if
	// layout[5] = 3, the 3rd key is mapped to the 5th symbol.

	switch mode {
	case "anneal":
		{

			msgChan := msg.MainChannel
			var wgPrint sync.WaitGroup
			var wgWork sync.WaitGroup
			wgPrint.Add(1)
			go msg.PrintMessages(msgChan, &wgPrint)

			seed := uint64(time.Now().UnixNano())
			for id := range runtime.NumCPU() {

				stream := uint64(id)
				source := rand.NewPCG(seed, stream)
				r := rand.New(source)

				r.Shuffle(nSym, func(i, j int) {
					layout[i], layout[j] = layout[j], layout[i]
				})

				in := anneal.AnnealInputs{
					Layout:     layout,
					Keys:       keys,
					MonoFreqs:  monoFreqs,
					BiFreqs:    biFreqs,
					Trigrams:   trigrams,
					SymToTrigs: symToTrigs,
				}

				wgWork.Add(1)
				go anneal.RunAnnealing(in, id, r, &wgWork)
			}

			wgWork.Wait()
			close(msg.MainChannel)
			wgPrint.Wait()
		}
	case "bruteforce": // W.I.P.
		{
		}
	}
}
