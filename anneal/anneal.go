package anneal

import (
	"sync"

	"github.com/aaa308-aub/lily58-anneal/assets"
	msg "github.com/aaa308-aub/lily58-anneal/assets/messages"
	cfg "github.com/aaa308-aub/lily58-anneal/config"

	"math"
	"math/bits"
	"math/rand/v2"
)

const (
	nSym     = cfg.NumSymbols // Assumed equal to number of included keys.
	nTrigram = cfg.NumTopTrigrams
	noTrigs  = cfg.IgnoreTrigrams
	nSteps   = cfg.NumAnnealSteps
)

var msgChan = msg.MainChannel
var sfbCosts = cfg.SFBCosts
var maxStretchesSq = cfg.MaxStretchesSq
var trigRewards = cfg.TrigramRewards

type keyT = cfg.KeyT
type trigramT = assets.TrigramT

// Purpose: deep-copying data to prevent cache contention between goroutines.
type AnnealInputs struct {
	Layout     [nSym]int
	Keys       [nSym]keyT
	MonoFreqs  [nSym]float32
	BiFreqs    [nSym][nSym]float32
	Trigrams   [nTrigram]trigramT
	SymToTrigs [nSym]uint64
}

// Excluded keys are not part of this LUT.
var distsSq = func() [nSym][nSym]float32 {

	// Filter out excluded keys first.
	var keys [nSym]keyT
	for i, j := 0, 0; i < cfg.NumKeysAll; i++ {
		key := cfg.KeysAll[i]
		if key.Fin != cfg.FingerNil {
			keys[j] = key
			j++
		}
	}

	var lut [nSym][nSym]float32
	for i := range nSym {
		for j := range nSym {
			key1 := keys[i]
			key2 := keys[j]
			lut[i][j] = distanceSquared(key1.X, key1.Y, key2.X, key2.Y)
		}
	}
	return lut
}()

// LUT to correct Schraudolph's fastExp function which has a
// periodic error function with period = ln(2). We use that to
// our advantage to sample over the period.
var fastExpErrs = func() [64]float64 {

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

	idx := (bits >> 46) & 0x3F // This function gets only weirder, huh?
	return exp * fastExpErrs[idx]
}

// Finds monogram cost of input symbol.
func monogramCost(
	symIdx int,
	freqs *[nSym]float32,
	layout *[nSym]int,
	keys *[nSym]keyT,
) float32 {

	key := keys[layout[symIdx]]
	weight := key.W
	return freqs[symIdx] * weight
}

// If modified, make sure to manually inline in bigramsCostWithSymbol function.
func bigramCost(
	symIdx1, symIdx2 int,
	freqs *[nSym][nSym]float32,
	layout *[nSym]int,
	keys *[nSym]keyT,
) float32 {

	keyIdx1, keyIdx2 := layout[symIdx1], layout[symIdx2]
	f1, f2 := keys[keyIdx1].Fin, keys[keyIdx2].Fin

	cost := sfbCosts[f1][f2]

	distSq := distsSq[keyIdx1][keyIdx2]
	maxStretchSq := maxStretchesSq[f1][f2]

	if distSq > maxStretchSq {
		cost *= (distSq - maxStretchSq) * cfg.StretchCostScaler
		// Notice: the above equation is not linear w.r.t. actual distances.
	}

	return cost * freqs[symIdx1][symIdx2]
}

// Finds sum of costs of all bigrams containing input symbol.
func bigramsCostWithSymbol(
	symIdx1 int,
	freqs *[nSym][nSym]float32,
	layout *[nSym]int,
	keys *[nSym]keyT,
) float32 {

	total := float32(0)

	// Note: Had to inline bigramCost myself in this for-loop because the
	// compiler refuses to. Without duplicating the code, execution time
	// grows wastefully by 20%.
	for symIdx2 := range nSym {
		keyIdx1, keyIdx2 := layout[symIdx1], layout[symIdx2]
		f1, f2 := keys[keyIdx1].Fin, keys[keyIdx2].Fin

		cost := sfbCosts[f1][f2]

		distSq := distsSq[keyIdx1][keyIdx2]
		maxStretchSq := maxStretchesSq[f1][f2]

		if distSq > maxStretchSq {
			cost *= (distSq - maxStretchSq) * cfg.StretchCostScaler
			// Notice: the above equation is not linear w.r.t. actual distances.
		}

		total += cost * freqs[symIdx1][symIdx2]
	}

	return total
}

// For every set bit (i.e., bit == 1) in input bitmask, finds the reward
// of the trigram in trigramInfos with index equal to bit position. The
// rewards are added together in absolute value, so other functions are
// expected to subtract this value.
func trigramsReward(
	bitmask uint64,
	trigrams *[nTrigram]trigramT,
	layout *[nSym]int,
	keys *[nSym]keyT,
) float32 {

	var total float32
	for ; bitmask != 0; bitmask &= (bitmask - 1) {
		idx := bits.TrailingZeros64(bitmask)
		t := &trigrams[idx]

		f1 := keys[layout[t.Syms[0]]].Fin
		f2 := keys[layout[t.Syms[1]]].Fin
		f3 := keys[layout[t.Syms[2]]].Fin

		reward := trigRewards[f1][f2][f3]
		total += t.Freq * reward
	}
	return total
}

func WholeLayoutCost(
	p *AnnealInputs,
) float32 {

	total := float32(0)

	for i := range p.Layout {
		total += monogramCost(i, &p.MonoFreqs, &p.Layout, &p.Keys)
	}

	for i := 1; i < nSym; i++ {
		for j := 0; j < i; j++ {
			total += bigramCost(i, j, &p.BiFreqs, &p.Layout, &p.Keys)
		}
	}

	if !noTrigs {
		const bitmaskAll = uint64(0xFF_FF_FF_FF_FF_FF_FF_FF)
		total -= trigramsReward(bitmaskAll, &p.Trigrams, &p.Layout, &p.Keys)
	}

	return total
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
	p *AnnealInputs,
) float32 {

	total := monogramCost(idx1, &p.MonoFreqs, &p.Layout, &p.Keys) +
		monogramCost(idx2, &p.MonoFreqs, &p.Layout, &p.Keys) +
		bigramsCostWithSymbol(idx1, &p.BiFreqs, &p.Layout, &p.Keys) +
		bigramsCostWithSymbol(idx2, &p.BiFreqs, &p.Layout, &p.Keys) -
		bigramCost(idx1, idx2, &p.BiFreqs, &p.Layout, &p.Keys)

	if noTrigs {
		return total
	}

	bitmaskUnion := p.SymToTrigs[idx1] | p.SymToTrigs[idx2]
	tr := trigramsReward(bitmaskUnion, &p.Trigrams, &p.Layout, &p.Keys)
	return total - tr
}

// Finds the right initial and final temperatures as well as the cooling
// factor, such that the average bad swap has a 90% acceptance chance
// initially and becomes 0.1% once 99% of swaps are processed.
//
// Notice: not thread-safe.
func findTempParameters(
	p *AnnealInputs,
	r *rand.Rand,
) (float64, float64, float64) {

	// Must deep-copy the layout in case it's needed un-modified and
	// restore at the end.
	layoutCopy := p.Layout

	deltaCostAvg := float32(0)

	// Tested different sample sizes and the average seems to converge
	// quickly for 1'000 bad swaps, so 10'000 is sufficient.
	for badSwapsCount := 0; badSwapsCount < 10000; {
		i := r.IntN(nSym)
		j := r.IntN(nSym)

		for i == j {
			// Whichever is re-rolled arguably doesn't matter so
			// this may be redundant.
			switch r.IntN(2) {
			case 0:
				i = r.IntN(nSym)
			case 1:
				j = r.IntN(nSym)
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

	tempI := float64(deltaCostAvg) / -math.Log(0.9)
	tempF := float64(deltaCostAvg) / -math.Log(0.001)

	const onePercentPoint = float64(nSteps * 99 / 100)
	coolingFactor := math.Pow(tempF/tempI, 1/onePercentPoint)

	return tempI, tempF, coolingFactor
}

// Be careful to input the worker goroutine's waitgroup, not the printer/logger.
func RunAnnealing(
	p AnnealInputs,
	id int,
	r *rand.Rand,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	scoreI := WholeLayoutCost(&p)
	msgChan <- msg.ThreadMessageT{
		ID:  id,
		Msg: msg.FormatInitialMessage(scoreI),
	}

	temp, _, coolingFactor := findTempParameters(&p, r)
	for range nSteps {
		i := r.IntN(nSym)
		j := r.IntN(nSym)

		for i == j {
			// Again, this may be redundant.
			switch r.IntN(2) {
			case 0:
				i = r.IntN(nSym)
			case 1:
				j = r.IntN(nSym)
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

	scoreF := WholeLayoutCost(&p)
	msgChan <- msg.ThreadMessageT{
		ID:  id,
		Msg: msg.FormatFinalMessage(scoreF, &p.Layout),
	}
}

// Used for remarkable layouts to dig deeper if possible.
func ProbeRegionGreedy(
	p AnnealInputs,
	id int,
	r *rand.Rand,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	scoreI := WholeLayoutCost(&p)
	msgChan <- msg.ThreadMessageT{
		ID:  id,
		Msg: msg.FormatInitialMessage(scoreI),
	}

	// Most of these steps are probably a waste of compute.
	for range nSteps {
		i := r.IntN(nSym)
		j := r.IntN(nSym)

		for i == j {
			// Again, this may be redundant.
			switch r.IntN(2) {
			case 0:
				i = r.IntN(nSym)
			case 1:
				j = r.IntN(nSym)
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

	scoreF := WholeLayoutCost(&p)
	msgChan <- msg.ThreadMessageT{
		ID:  id,
		Msg: msg.FormatFinalMessage(scoreF, &p.Layout),
	}
}
