package annealing

import (
	"github.com/aaa308-aub/lily58-anneal/assets"
	cf "github.com/aaa308-aub/lily58-anneal/config"
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
	symbolToTrigrams *[numSymbols * numTopTrigrams]int8
}

// Builder function used to pass an instance of annealingParams to other
// functions without having to export this type of struct.
func CoupleAnnealingParams(
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
	monogramFreqs *[numSymbols]float32,
	bigramFreqs *[numSymbols * numSymbols]float32,
	trigramInfos *[numTopTrigrams]trigramInfo,
	symbolToTrigrams *[numSymbols * numTopTrigrams]int8,
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

func distanceSquared(x1, y1, x2, y2 float32) float32 {

	dx, dy := x1-x2, y1-y2
	return dx*dx + dy*dy
}

func monogramCost(freq, keyWeight float32) float32 {
	return freq * keyWeight
}

func bigramCost(
	freq float32,
	finger1, finger2 finger,
	distanceSquared float32,
	stretchLimitSquared float32,
) float32 {

	cost := float32(0)

	// Putting my faith in the branch predictor here. Will probably
	// replace with another LUT.
	if finger1 == finger2 {
		cost += cf.PenaltySFB
	}

	if distanceSquared > stretchLimitSquared {
		cost += (distanceSquared - stretchLimitSquared) * cf.PenaltyStretch
	}

	return cost * freq
}

// Returns the reward in absolute value. Other cost functions are expected
// to subtract this value, not add it.
func trigramReward(
	freq float32,
	orderedFingers *[3]finger,
) float32 {

	// TODO: Replace with LUT. I don't think Go will make its own.
	switch *orderedFingers {
	case [3]finger{
		cf.FingerRing,
		cf.FingerMiddle,
		cf.FingerIndex}:
		return cf.RewardInwardRoll * freq

	case [3]finger{
		cf.FingerIndex,
		cf.FingerMiddle,
		cf.FingerRing}:
		return cf.RewardOutwardRoll * freq // Typically a lesser reward.

	default:
		return 0
	}
}

func InitialLayoutCost(
	p annealingParams,
) float32 {

	// To have an easier time reading this function, remember that the layout
	// is a fixed array of keys and the symbol-indices are scattered across.

	cost := float32(0)

	for key1, symbol1 := range p.layout {
		// Cost of monogram.
		cost += monogramCost(
			p.monogramFreqs[symbol1],
			p.keyInfos[key1].Weight,
		)

		// Cost of bigrams.
		offset := numSymbols * symbol1
		for key2 := 0; key2 < len(p.layout); key2++ {

			symbol2 := p.layout[key2]
			bigram := offset + symbol2

			key1Info, key2Info := p.keyInfos[key1], p.keyInfos[key2]

			finger1 := key1Info.AssignedFinger
			finger2 := key2Info.AssignedFinger

			distanceSquared := distanceSquared(
				key1Info.X, key1Info.Y,
				key2Info.X, key2Info.Y,
			)

			stretchLimitSquared := cf.StretchLimitsSquared[finger1*numFingers+finger2]

			cost += bigramCost(
				p.bigramFreqs[bigram],
				finger1, finger2,
				distanceSquared,
				stretchLimitSquared,
			)
		}
	}

	// Reward of trigrams.
	symbolToKey := make(map[int]int, numSymbols)
	for key, symbol := range p.layout {
		symbolToKey[symbol] = key
	}

	var orderedFingers [3]finger
	for _, trigram := range p.trigramInfos {
		for i, symbol := range trigram.OrderedSymbols {
			key := symbolToKey[int(symbol)]
			orderedFingers[i] = p.keyInfos[key].AssignedFinger
		}

		cost -= trigramReward(trigram.Freq, &orderedFingers)
	}

	return cost
}

// Finds the cost contributed by the symbol occupying layout[key].
func symbolContribution(
	key1 int,
	p annealingParams,
	isWithTrigrams bool,
) float32 {
	symbol1 := p.layout[key1]
	cost := monogramCost(p.monogramFreqs[symbol1], p.keyInfos[key1].Weight)

	// Cost of bigrams.
	offset := numSymbols * symbol1
	for key2 := 0; key2 < len(p.layout); key2++ {

		symbol2 := p.layout[key2]
		bigram := offset + symbol2

		key1Info, key2Info := p.keyInfos[key1], p.keyInfos[key2]

		finger1 := key1Info.AssignedFinger
		finger2 := key2Info.AssignedFinger

		distanceSquared := distanceSquared(
			key1Info.X, key1Info.Y,
			key2Info.X, key2Info.Y,
		)

		stretchLimitSquared := cf.StretchLimitsSquared[finger1*numFingers+finger2]

		cost += bigramCost(
			p.bigramFreqs[bigram],
			finger1, finger2,
			distanceSquared,
			stretchLimitSquared,
		)
	}

	if !isWithTrigrams {
		return cost
	}

	var orderedFingers [3]finger
	//offset = numSymbols * symbol1
	for i := offset; i < offset+numTopTrigrams; i++ {

		// TODO: Find a way to remove this branch. Probably another LUT.
		trigramIndex := p.symbolToTrigrams[i]
		switch trigramIndex {
		case -1:
			continue
		default:
			{
				trigram := p.trigramInfos[trigramIndex]
				for i, symbol := range trigram.OrderedSymbols {

					// TODO: Replace this temporary linear search.
					for key := range p.layout {
						if p.layout[key] == int(symbol) {
							orderedFingers[i] = p.keyInfos[key].AssignedFinger
						}
					}
				}

				cost -= trigramReward(trigram.Freq, &orderedFingers)
			}
		}
	}

	return cost
}
