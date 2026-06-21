package annealing

import (
	"github.com/aaa308-aub/lily58-anneal/assets"
	cf "github.com/aaa308-aub/lily58-anneal/config"

	"math/bits"
)

const numSymbols = len(cf.TargetSymbols) // Assumed equal to number of included keys.
const numTopTrigrams = cf.NumTopTrigrams
const numFingers = cf.NumFingers

type finger = cf.Finger
type keyInfo = cf.KeyInfo
type trigramInfo = assets.TrigramInfo

type annealingParams struct {
	layout           *[numSymbols]int
	keyInfos         *[numSymbols]keyInfo
	monogramFreqs    *[numSymbols]float32
	bigramFreqs      *[numSymbols * numSymbols]float32
	trigramInfos     *[numTopTrigrams]trigramInfo
	symbolToTrigrams *[numSymbols]uint64
}

// Builder function used to pass an instance of annealingParams to other
// functions without having to export this type of struct.
func CoupleAnnealingParams(
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
	monogramFreqs *[numSymbols]float32,
	bigramFreqs *[numSymbols * numSymbols]float32,
	trigramInfos *[numTopTrigrams]trigramInfo,
	symbolToTrigrams *[numSymbols]uint64,
) annealingParams {

	return annealingParams{
		layout,
		keyInfos,
		monogramFreqs,
		bigramFreqs,
		trigramInfos,
		symbolToTrigrams,
	}
}

// Distances are squared to avoid expensive square-root functions.
func distanceSquared(x1, y1, x2, y2 float32) float32 {

	dx, dy := x1-x2, y1-y2
	return dx*dx + dy*dy
}

// Finds monogram cost of input symbol. Fetches required info on its own.
func monogramCost(
	symbolIdx int,
	freqs *[numSymbols]float32,
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
) float32 {

	key := keyInfos[layout[symbolIdx]]
	weight := key.Weight
	return freqs[symbolIdx] * weight
}

// Finds sum of costs of all bigrams containing input symbol. Fetches
// required info.
func bigramsCost(
	symbolIdx1 int,
	freqs *[numSymbols * numSymbols]float32,
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
) float32 {

	totalCost := float32(0)
	key1 := keyInfos[layout[symbolIdx1]]
	finger1, x1, y1 := key1.AssignedFinger, key1.X, key1.Y

	offset := numSymbols * symbolIdx1
	for symbolIdx2 := 0; symbolIdx2 < numSymbols; symbolIdx2++ {
		cost := float32(0)
		key2 := keyInfos[layout[symbolIdx2]]
		finger2, x2, y2 := key2.AssignedFinger, key2.X, key2.Y

		// Punish same-finger bigrams.
		cost += cf.BigramFingersPenalty[finger1*numFingers+finger2]

		// We could make LUTs for this, but for now I prefer this.
		distanceSq := distanceSquared(x1, y1, x2, y2)
		stretchLimitSq := cf.StretchLimitsSquared[finger1*numFingers+finger2]

		if distanceSq > stretchLimitSq {
			cost *= (distanceSq - stretchLimitSq) * cf.PenaltyStretchScaler
			// Notice: the above equation is not linear w.r.t. actual distances.
		}

		totalCost += cost * freqs[offset+symbolIdx2]
	}

	return totalCost
}

// For every set bit (i.e., bit == 1) in input bitmask, finds the reward
// of the trigram in trigramInfos with index equal to bit position. The
// rewards are added together in absolute value, so other functions are
// expected to subtract this value.
func trigramsReward(
	bitmask uint64,
	trigramInfos *[numTopTrigrams]trigramInfo,
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
) float32 {

	var totalCost float32
	for ; bitmask != 0; bitmask &= (bitmask - 1) {
		idx := bits.TrailingZeros64(bitmask)
		t := &trigramInfos[idx]

		f1 := keyInfos[layout[t.OrderedSymbols[0]]].AssignedFinger
		f2 := keyInfos[layout[t.OrderedSymbols[1]]].AssignedFinger
		f3 := keyInfos[layout[t.OrderedSymbols[2]]].AssignedFinger

		reward := cf.TrigramFingersReward[f1*numFingers*numFingers+f2*numFingers+f3]
		totalCost += t.Freq * reward
	}
	return totalCost
}

func InitialLayoutCost(
	p annealingParams, // symbolToTrigrams is unused but that's okay.
) float32 {

	cost := float32(0)

	for symbolIdx := range p.layout {
		// Cost of monogram.
		cost += monogramCost(symbolIdx, p.monogramFreqs, p.layout, p.keyInfos)

		// Cost of bigrams.
		cost += bigramsCost(symbolIdx, p.bigramFreqs, p.layout, p.keyInfos)
	}

	// Reward of trigrams.
	const bitmaskFull = uint64(0xFF_FF_FF_FF_FF_FF_FF_FF)
	cost -= trigramsReward(bitmaskFull, p.trigramInfos, p.layout, p.keyInfos)

	return cost
}

// Finds the cost contributed by two symbols.
//
// In the context of our engine, the cost contributed by one symbol is never
// calculated alone, and we want to compute the union of bitmasks for trigrams
// to avoid double-counting those that contain both swapped symbols. So this
// is technically more idiomatic than implementing a single symbolContribution
// function. Both are clean, though.
func twoSymbolsContribution(
	idx1, idx2 int,
	p annealingParams,
) float32 {
	bitmaskUnion := p.symbolToTrigrams[idx1] | p.symbolToTrigrams[idx2]

	return monogramCost(idx1, p.monogramFreqs, p.layout, p.keyInfos) +
		monogramCost(idx2, p.monogramFreqs, p.layout, p.keyInfos) +
		bigramsCost(idx1, p.bigramFreqs, p.layout, p.keyInfos) +
		bigramsCost(idx2, p.bigramFreqs, p.layout, p.keyInfos) -
		trigramsReward(bitmaskUnion, p.trigramInfos, p.layout, p.keyInfos)
}
