package config

import (
	"fmt"
	"unicode"
)

const (
	IgnoreTrigrams    = false // true | false.
	ExcludedKeySymbol = '·'
	NumAnnealSteps    = 50_000_000
	NumKeysAll        = 29
	NumSymbols        = len(symbolsStr)
	NumTopTrigrams    = int8(64)
	StretchCostScaler = float32(0.5) // Do not raise too aggressively.
)

// 2D LUT for penalty of fingers used to type bigram. Punishes
// same-finger bigrams (SFBs) only by default.
var SFBCosts = [NumFingers][NumFingers]float32{
	{2, 1, 1},
	{1, 2, 1},
	{1, 1, 2},
}

// 2D LUT for the max stretch distance allowed for bigrams not to
// be punished, given two fingers. The values are squared to
// prevent wasting cycles on math.Sqrt function. Matrix must be
// symmetric because order of fingers/symbols in bigrams doesn't
// matter.
var MaxStretchesSq = [NumFingers][NumFingers]float32{
	{0, 1.56, 9},
	{1.56, 0, 6.25},
	{9, 6.25, 0},
}

// 3D LUT for trigram inward roll (ring->middle->index) reward
// multiplier. Other distinct-finger trigrams are also rewarded slightly.
var TrigramRewards = func() [NumFingers][NumFingers][NumFingers]float32 {

	var lut [NumFingers][NumFingers][NumFingers]float32
	lut[FingerRing][FingerMiddle][FingerIndex] = 5   // R->M->I
	lut[FingerIndex][FingerMiddle][FingerRing] = 3.5 // I->M->R
	lut[FingerMiddle][FingerRing][FingerIndex] = 3.5 // M->R->I
	lut[FingerMiddle][FingerIndex][FingerRing] = 3.5 // M->I->R
	return lut
}()

type FingerT uint8

const (
	FingerRing   FingerT = iota // 0
	FingerMiddle                // 1
	FingerIndex                 // 2
	NumFingers                  // 3
	FingerNil                   // 4
	// Typically the null-value is the zero-value, but there are good
	// reasons to make an exception here, for LUTs involving fingers.
)

// X, Y: coordinates, W: Key weight, F: Finger assigned.
type KeyT struct {
	X, Y, W float32
	Fin     FingerT
}

// Do not touch above this line unless you know what you're doing.
//
// To exclude a key (make the engine ignore it), assign it to FingerNil.
// To re-include it, assign it to a valid finger.
//
// Note: the key with coordinates 4x, -0.58y (see images in doc) is
// considered Row 4, Column 4 for simplicity.
//
// If you want to adjust a key's weight, change its 3rd field.
// For example, the weight of {0, 2, 1.5, FingerMiddle} is 1.5.

var KeysAll = [NumKeysAll]KeyT{
	// Row 0
	{-2, 1.67, 2.5, FingerMiddle}, {-1, 1.76, 2, FingerMiddle}, {0, 2, 1.5, FingerMiddle}, {1, 2.08, 2, FingerIndex}, {2, 2, 2.5, FingerIndex}, {3, 1.92, 3, FingerIndex},
	// Row 1
	{-2, 0.67, 2, FingerRing}, {-1, 0.76, 1.5, FingerRing}, {0, 1, 1, FingerMiddle}, {1, 1.08, 1.5, FingerIndex}, {2, 1, 2, FingerIndex}, {3, 0.92, 2.5, FingerIndex},
	// Row 2
	{-2, -0.33, 1.5, FingerRing}, {-1, -0.24, 1, FingerRing}, {0, 0, 1.5, FingerMiddle}, {1, 0.08, 1, FingerIndex}, {2, 0, 1.5, FingerIndex}, {3, -0.08, 2.5, FingerIndex},
	// Row 3 -- key reserved for pinky excluded by default
	{-2, -1.33, 2, FingerNil}, {-1, -1.24, 1.5, FingerRing}, {0, -1, 2, FingerMiddle}, {1, -0.92, 1.5, FingerIndex}, {2, -1, 2, FingerIndex}, {3, -1.08, 2.5, FingerIndex},
	// Row 4 -- keys reserved for thumb excluded by default
	{0.5, -2, 2.5, FingerMiddle}, {1.5, -2, 2.5, FingerIndex}, {2.5, -2.08, 3, FingerNil}, {3.8, -2.09, 3.5, FingerNil}, {4, -0.58, 3.5, FingerIndex},

	// Row/Column guide:                              Blank layout (scratch):
	//
	// [0,0][0,1][0,2][0,3][0,4][0,5]                   [][][][][][]
	// [1,0][1,1][1,2][1,3][1,4][1,5]                   [][][][][][]
	// [2,0][2,1][2,2][2,3][2,4][2,5] [4,4]             [][][][][][] []
	// [3,0][3,1][3,2][3,3][3,4][3,5]                   [][][][][][]
	//             [4,0][4,1][4,2][4,3]                      [][][][]
}

// Place your symbols you want mapped below. Please make sure the symbols
// belong to the language you chose, and that their number is the same as
// the number of keys included. Otherwise, an error will occur.

const symbolsStr = "abcdefghijklmnopqrstuvwxyz"

// For the available languages and their alphabets, see the README.

const TargetLanguageCode = "en"

// Don't touch below this line unless you know what you're doing.

var SymbolsArr = func() [NumSymbols]rune {
	var arr [NumSymbols]rune
	for i, s := range symbolsStr {
		arr[i] = unicode.ToLower(s)
	}
	return arr
}()

var KeysIncluded = func() [NumSymbols]KeyT {

	nKeysIncluded := 0
	for _, key := range KeysAll {
		if key.Fin != FingerNil {
			nKeysIncluded++
		}
	}

	if nKeysIncluded < 3 || nKeysIncluded > 29 {
		panic(fmt.Errorf(
			"number of included keys (%d) must be between 3 and 29 inclusive",
			nKeysIncluded,
		))
	}

	if nKeysIncluded != NumSymbols {
		panic(fmt.Errorf(
			"number of included keys (%d) is different from number of symbols (%d)",
			nKeysIncluded,
			NumSymbols,
		))
	}

	seen := make(map[KeyT]struct{}, NumSymbols)

	var keys [NumSymbols]KeyT
	for i, j := 0, 0; i < NumKeysAll; i++ {
		key := KeysAll[i]
		if key.Fin != FingerNil {

			_, ok := seen[key]
			if ok {
				panic("found two keys in config with identical fields")
			}

			keys[j] = key
			seen[key] = struct{}{}
			j++
		}
	}
	return keys
}()
