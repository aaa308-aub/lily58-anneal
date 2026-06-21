package annealing

import (
	"fmt"
	"sync"

	"github.com/aaa308-aub/lily58-anneal/assets"
	cf "github.com/aaa308-aub/lily58-anneal/config"

	"math"
	"math/bits"
	"math/rand/v2"
)

const numSymbols = len(cf.TargetSymbols) // Assumed equal to number of included keys.
const numTopTrigrams = cf.NumTopTrigrams
const numFingers = cf.NumFingers

type finger = cf.Finger
type keyInfo = cf.KeyInfo
type trigramInfo = assets.TrigramInfo

type AnnealingInputs struct { // Deep copy so that goroutines don't fight for data.
	Layout           [numSymbols]int
	KeyInfos         [numSymbols]keyInfo
	MonogramFreqs    [numSymbols]float32
	BigramFreqs      [numSymbols * numSymbols]float32
	TrigramInfos     [numTopTrigrams]trigramInfo
	SymbolToTrigrams [numSymbols]uint64
}

// Distances are squared to avoid expensive square-root functions.
func distanceSquared(x1, y1, x2, y2 float32) float32 {

	dx, dy := x1-x2, y1-y2
	return dx*dx + dy*dy
}

// This is the algorithm she told you not to worry about.
//
// Credit: Nicol N. Schraudolph's fast approximation for e^x.
func fastExp(x float64) float64 {
	bits := int64(6497320848556798*x + 4607182418800017408)
	exp := math.Float64frombits(uint64(bits))
	return exp
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

func wholeLayoutCost(
	p *AnnealingInputs,
) float32 {

	cost := float32(0)

	for symbolIdx := range p.Layout {
		// Cost of monogram.
		cost += monogramCost(symbolIdx, &p.MonogramFreqs, &p.Layout, &p.KeyInfos)

		// Cost of bigrams.
		cost += bigramsCost(symbolIdx, &p.BigramFreqs, &p.Layout, &p.KeyInfos)
	}

	// Reward of trigrams.
	const bitmaskFull = uint64(0xFF_FF_FF_FF_FF_FF_FF_FF)
	cost -= trigramsReward(bitmaskFull, &p.TrigramInfos, &p.Layout, &p.KeyInfos)

	return cost
}

// Finds the cost contributed by two symbols.
//
// In the context of our engine, the cost contributed by one symbol is never
// calculated alone, and we want to compute the union of bitmasks for trigrams
// to avoid double-counting those that contain both swapped symbols. So this
// is technically more idiomatic than implementing a single symbolContribution
// function. Neither implementation is drastically better.
func twoSymbolsContribution(
	idx1, idx2 int,
	p *AnnealingInputs,
) float32 {

	bitmaskUnion := p.SymbolToTrigrams[idx1] | p.SymbolToTrigrams[idx2]

	return monogramCost(idx1, &p.MonogramFreqs, &p.Layout, &p.KeyInfos) +
		monogramCost(idx2, &p.MonogramFreqs, &p.Layout, &p.KeyInfos) +
		bigramsCost(idx1, &p.BigramFreqs, &p.Layout, &p.KeyInfos) +
		bigramsCost(idx2, &p.BigramFreqs, &p.Layout, &p.KeyInfos) -
		trigramsReward(bitmaskUnion, &p.TrigramInfos, &p.Layout, &p.KeyInfos)
}

// Finds the right initial and final temperatures as well as the cooling
// factor, such that the average bad swap has a 90% acceptance chance
// initially and becomes 0.1% once 99% of swaps are processed.
func findTempParameters(
	p *AnnealingInputs,
	r *rand.Rand,
) (float64, float64, float64) {

	deltaCostAvg := float32(0)

	// Tested different sample sizes and the average seems to converge
	// quickly for 1'000 bad swaps, so 10'000 is sufficient.
	for badSwapsCounter := 0; badSwapsCounter < 10000; {
		i := r.IntN(numSymbols)
		j := r.IntN(numSymbols)

		for i == j {
			// Whichever is re-rolled arguably doesn't matter so
			// this may be redundant.
			switch r.IntN(2) {
			case 0:
				i = r.IntN(numSymbols)
			case 1:
				j = r.IntN(numSymbols)
			}
		}

		costOld := twoSymbolsContribution(i, j, p)
		p.Layout[i], p.Layout[j] = p.Layout[j], p.Layout[i]
		costNew := twoSymbolsContribution(i, j, p)

		deltaCost := costNew - costOld
		if deltaCost > 0 {
			badSwapsCounter++
			deltaCostAvg += (deltaCost - deltaCostAvg) / float32(badSwapsCounter)
		}
	}

	tempInitial := float64(deltaCostAvg) / -math.Log(0.9)
	tempFinal := float64(deltaCostAvg) / -math.Log(0.001)

	const OnePercentThreshold = float64(cf.NumAnnealingSteps * 99 / 100)
	coolingFactor := math.Pow(tempFinal/tempInitial, 1/OnePercentThreshold)

	return tempInitial, tempFinal, coolingFactor
}

func RunAnnealing(
	p AnnealingInputs,
	r *rand.Rand,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	temp, _, coolingFactor := findTempParameters(&p, r)

	score := wholeLayoutCost(&p)
	fmt.Printf("initial score: %f\n", score)

	for i := 0; i < cf.NumAnnealingSteps; i++ {
		i := r.IntN(numSymbols)
		j := r.IntN(numSymbols)

		for i == j {
			// Again, this may be redundant.
			switch r.IntN(2) {
			case 0:
				i = r.IntN(numSymbols)
			case 1:
				j = r.IntN(numSymbols)
			}
		}

		costOld := twoSymbolsContribution(i, j, &p)
		p.Layout[i], p.Layout[j] = p.Layout[j], p.Layout[i]
		costNew := twoSymbolsContribution(i, j, &p)

		deltaCost := costNew - costOld
		if deltaCost <= 0 || r.Float64() <= fastExp(-float64(deltaCost)/temp) {
			score += deltaCost
		} else { // Rejected; switch back.
			p.Layout[i], p.Layout[j] = p.Layout[j], p.Layout[i]
		}

		temp *= coolingFactor
	}

	fmt.Printf("final score: %f\t", score)
	fmt.Printf("|\tlayout: %v\n", p.Layout)
}
