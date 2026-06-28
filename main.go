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

func main() {

	const nSym = cfg.NumSymbols

	// nSym == nKeysIncluded is checked within config.go which will panic if not.

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
				"found symbol (%q) with no monogram data, may be a duplicate or "+
					"may not belong to target language",
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

	var keys = cfg.KeysIncluded

	var layout [nSym]int // equals [0, 1, 2... numSymbols-1].
	for i := range layout {
		layout[i] = i
	}

	// Note: A layout could mean either a mapping of keys to symbols or a
	// mapping of symbols to keys (it's a bijection). So to be clear, if
	// layout[5] = 3, the 3rd key is mapped to the 5th symbol.

	msgChan := make(chan msg.ThreadMessageT, 100)
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
		go anneal.RunAnnealing(in, r, id, msgChan, &wgWork)
	}

	wgWork.Wait()
	close(msgChan)
	wgPrint.Wait()
}
