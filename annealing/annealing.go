package annealing

import (
	"fmt"
	"sync"

	"github.com/aaa308-aub/lily58-anneal/assets"
	msg "github.com/aaa308-aub/lily58-anneal/assets/messages"
	cfg "github.com/aaa308-aub/lily58-anneal/config"

	"math"
	"math/bits"
	"math/rand/v2"
)

const (
	numSymbols     = cfg.NumSymbols // Assumed equal to number of included keys.
	numTopTrigrams = cfg.NumTopTrigrams
	numFingers     = cfg.NumFingers
	ignoreTrigrams = cfg.IgnoreTrigrams
)

var msgChan = msg.MainChannel

type keyInfo = cfg.KeyInfo
type trigramInfo = assets.TrigramInfo

// Purpose: deep-copying data to prevent cache contention between goroutines.
type AnnealingInputs struct {
	Layout           [numSymbols]int
	KeyInfos         [numSymbols]keyInfo
	MonogramFreqs    [numSymbols]float32
	BigramFreqs      [numSymbols * numSymbols]float32
	TrigramInfos     [numTopTrigrams]trigramInfo
	SymbolToTrigrams [numSymbols]uint64
}

// Excluded keys are not part of this LUT.
var distancesSquaredLUT = func() [numSymbols * numSymbols]float32 {

	// Filter out excluded keys first.
	var keys [numSymbols]cfg.KeyInfo
	for i, j := 0, 0; i < cfg.NumKeys; i++ {
		key := cfg.KeyInfos[i]
		if key.AssignedFinger != cfg.FingerNil {
			keys[j] = key
			j++
		}
	}

	const N = numSymbols
	var lut [N * N]float32
	for i := range N {
		for j := range N {
			key1 := keys[i]
			key2 := keys[j]
			lut[N*i+j] = distanceSquared(key1.X, key1.Y, key2.X, key2.Y)
		}
	}
	return lut
}()

// LUT to correct Schraudolph's fastExp function which has a
// periodic error function with period = ln(2). We use that to
// our advantage to sample over the period.
var fastExpErrorLUT = func() [64]float64 {

	var lut [64]float64
	for i := range 64 {
		x := float64(i) * math.Log(2) / 64
		bits := int64(6497320848556798*x + 4607182418800017408)
		approx := math.Float64frombits(uint64(bits))
		lut[i] = math.Exp(x) / approx
	}
	return lut
}()

// Distances are squared to avoid expensive square-root functions.
func distanceSquared(x1, y1, x2, y2 float32) float32 {

	dx, dy := x1-x2, y1-y2
	return dx*dx + dy*dy
}

// This is the algorithm she told you not to worry about.
//
// Credit: Nicol N. Schraudolph's fast approximation for e^x, modified
// with error-correcting samples. Error margin went down from 6% to 0.17%.
//
// See special cases if |x| > ~708 -- would normally cause overflow.
func fastExp(x float64) float64 {

	if x > 700 {
		return math.Inf(1)
	}
	if x < -700 {
		return 0
	}

	bits := int64(6497320848556798*x + 4607182418800017408)
	exp := math.Float64frombits(uint64(bits))

	index := (bits >> 46) & 0x3F // This function gets only weirder, huh?
	return exp * fastExpErrorLUT[index]
}

// Finds monogram cost of input symbol.
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

func bigramCost(
	symbolIdx1, symbolIdx2 int,
	freqs *[numSymbols * numSymbols]float32,
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
) float32 {

	keyIdx1, keyIdx2 := layout[symbolIdx1], layout[symbolIdx2]
	finger1 := keyInfos[keyIdx1].AssignedFinger
	finger2 := keyInfos[keyIdx2].AssignedFinger

	cost := cfg.BigramFingersPenalty[finger1*numFingers+finger2]

	distanceSq := distancesSquaredLUT[numSymbols*keyIdx1+keyIdx2]
	stretchLimitSq := cfg.StretchLimitsSquared[numFingers*finger1+finger2]

	if distanceSq > stretchLimitSq {
		cost *= (distanceSq - stretchLimitSq) * cfg.PenaltyStretchScaler
		// Notice: the above equation is not linear w.r.t. actual distances.
	}

	return cost * freqs[numSymbols*symbolIdx1+symbolIdx2]
}

// Finds sum of costs of all bigrams containing input symbol.
func bigramsCostWithSymbol(
	symbolIdx1 int,
	freqs *[numSymbols * numSymbols]float32,
	layout *[numSymbols]int,
	keyInfos *[numSymbols]keyInfo,
) float32 {

	totalCost := float32(0)

	// Note: Had to inline bigramCost myself in this for-loop because the
	// compiler refuses to. Without duplicating the code, execution time
	// grows wastefully by 20%.
	for symbolIdx2 := range numSymbols {
		keyIdx1, keyIdx2 := layout[symbolIdx1], layout[symbolIdx2]
		finger1 := keyInfos[keyIdx1].AssignedFinger
		finger2 := keyInfos[keyIdx2].AssignedFinger

		cost := cfg.BigramFingersPenalty[finger1*numFingers+finger2]

		distanceSq := distancesSquaredLUT[numSymbols*keyIdx1+keyIdx2]
		stretchLimitSq := cfg.StretchLimitsSquared[numFingers*finger1+finger2]

		if distanceSq > stretchLimitSq {
			cost *= (distanceSq - stretchLimitSq) * cfg.PenaltyStretchScaler
		}

		totalCost += cost * freqs[numSymbols*symbolIdx1+symbolIdx2]
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

		reward := cfg.TrigramFingersReward[f1*numFingers*numFingers+f2*numFingers+f3]
		totalCost += t.Freq * reward
	}
	return totalCost
}

func WholeLayoutCost(
	p *AnnealingInputs,
) float32 {

	cost := float32(0)

	for symbolIdx := range p.Layout {
		// Cost of monogram.
		cost += monogramCost(symbolIdx, &p.MonogramFreqs, &p.Layout, &p.KeyInfos)
	}

	// Cost of bigrams.
	for symbolIdx1 := 1; symbolIdx1 < numSymbols; symbolIdx1++ {
		for symbolIdx2 := 0; symbolIdx2 < symbolIdx1; symbolIdx2++ {
			cost += bigramCost(symbolIdx1, symbolIdx2, &p.BigramFreqs, &p.Layout, &p.KeyInfos)
		}
	}

	// Reward of trigrams.
	if !ignoreTrigrams {
		const bitmaskFull = uint64(0xFF_FF_FF_FF_FF_FF_FF_FF)
		cost -= trigramsReward(bitmaskFull, &p.TrigramInfos, &p.Layout, &p.KeyInfos)
	}

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

	cost := monogramCost(idx1, &p.MonogramFreqs, &p.Layout, &p.KeyInfos) +
		monogramCost(idx2, &p.MonogramFreqs, &p.Layout, &p.KeyInfos) +
		bigramsCostWithSymbol(idx1, &p.BigramFreqs, &p.Layout, &p.KeyInfos) +
		bigramsCostWithSymbol(idx2, &p.BigramFreqs, &p.Layout, &p.KeyInfos) -
		bigramCost(idx1, idx2, &p.BigramFreqs, &p.Layout, &p.KeyInfos)

	if ignoreTrigrams {
		return cost
	}

	bitmaskUnion := p.SymbolToTrigrams[idx1] | p.SymbolToTrigrams[idx2]
	tr := trigramsReward(bitmaskUnion, &p.TrigramInfos, &p.Layout, &p.KeyInfos)
	return cost - tr
}

// Finds the right initial and final temperatures as well as the cooling
// factor, such that the average bad swap has a 90% acceptance chance
// initially and becomes 0.1% once 99% of swaps are processed.
//
// Notice: not thread safe.
func findTempParameters(
	p *AnnealingInputs,
	r *rand.Rand,
) (float64, float64, float64) {

	// Must deep-copy the layout in case it's needed un-modified and
	// restore at the end.
	layoutCopy := p.Layout

	deltaCostAvg := float32(0)

	// Tested different sample sizes and the average seems to converge
	// quickly for 1'000 bad swaps, so 10'000 is sufficient.
	for badSwapsCount := 0; badSwapsCount < 10000; {
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
			badSwapsCount++
			deltaCostAvg += (deltaCost - deltaCostAvg) / float32(badSwapsCount)
		}
	}

	p.Layout = layoutCopy

	tempInitial := float64(deltaCostAvg) / -math.Log(0.9)
	tempFinal := float64(deltaCostAvg) / -math.Log(0.001)

	const onePercentThreshold = float64(cfg.NumAnnealingSteps * 99 / 100)
	coolingFactor := math.Pow(tempFinal/tempInitial, 1/onePercentThreshold)

	return tempInitial, tempFinal, coolingFactor
}

// Be careful to input the worker goroutine's waitgroup, not the printer/logger.
func RunAnnealing(
	p AnnealingInputs,
	id int,
	r *rand.Rand,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	temp, _, coolingFactor := findTempParameters(&p, r)
	initialScore := WholeLayoutCost(&p)

	msg.MainChannel <- msg.ThreadMessage{
		ThreadID: id,
		Message: fmt.Sprintf(
			"Started SA with initial score of %f\n", initialScore,
		),
	}

	for range cfg.NumAnnealingSteps {
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
		if deltaCost > 0 && r.Float64() > fastExp(-float64(deltaCost)/temp) {
			// Rejected; switch back.
			p.Layout[i], p.Layout[j] = p.Layout[j], p.Layout[i]
		}

		temp *= coolingFactor
	}

	finalScore := WholeLayoutCost(&p)

	endMsg := msg.ThreadMessage{
		ThreadID: id,
		Message:  msg.FormatFinalMessage(finalScore, &p.Layout),
	}

	msg.MainChannel <- endMsg
}

// Used for remarkable layouts to dig deeper if possible.
func ProbeRegionGreedy(
	p AnnealingInputs,
	id int,
	r *rand.Rand,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	initialScore := WholeLayoutCost(&p)

	msg.MainChannel <- msg.ThreadMessage{
		ThreadID: id,
		Message: fmt.Sprintf(
			"Started SA with initial score of %f\n", initialScore,
		),
	}

	// Most of these steps are probably a waste of compute.
	for range cfg.NumAnnealingSteps {
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
		if deltaCost > 0 {
			// Rejected; switch back.
			p.Layout[i], p.Layout[j] = p.Layout[j], p.Layout[i]
		}
	}

	finalScore := WholeLayoutCost(&p)

	endMsg := msg.ThreadMessage{
		ThreadID: id,
		Message:  msg.FormatFinalMessage(finalScore, &p.Layout),
	}

	msg.MainChannel <- endMsg
}
