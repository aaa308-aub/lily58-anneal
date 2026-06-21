package config

const NumAnnealingSteps = 50_000_000
const NumKeys = 29
const NumTopTrigrams = int8(64)
const PenaltyStretchScaler = float32(0.5) // Do not raise too aggressively.

// Small 2D LUT for penalty of fingers used to type bigram. Used to
// only punish same-finger bigrams by default.
var BigramFingersPenalty = [NumFingers * NumFingers]float32{
	2, 1, 1,
	1, 2, 1,
	1, 1, 2,
}

// Small 2D LUT for the max stretch distance allowed for bigrams
// not to be punished, given two fingers. The values are squared to
// prevent wasting cycles on math.Sqrt function. Matrix must be
// symmetric because order of fingers/symbols in bigrams doesn't
// matter.
var StretchLimitsSquared = [NumFingers * NumFingers]float32{
	0, 1.56, 10.56,
	1.56, 0, 6.25,
	10.56, 6.25, 0,
}

// Small 3D LUT for trigram inward roll (ring->middle->index) reward
// multiplier, and a smaller one for outward rolls (index->middle->ring).
var TrigramFingersReward = [NumFingers * NumFingers * NumFingers]float32{
	0, 0, 0,
	0, 0, 7.5,
	0, 0, 0,
	0, 0, 0,
	0, 0, 0,
	0, 0, 0,
	0, 0, 0,
	2.5, 0, 0,
	0, 0, 0,
}

type Finger uint8

const (
	FingerRing   Finger = iota // 0
	FingerMiddle               // 1
	FingerIndex                // 2
	NumFingers                 // 3
	FingerNil                  // 4
	// Typically the null-value is the zero-value, but there are good
	// reasons to make an exception here, for LUTs involving fingers.
)

type KeyInfo struct {
	X, Y, Weight   float32
	AssignedFinger Finger
}

// Do not touch above this line unless you know what you're doing.
//
// To exclude a key (make the engine ignore it), assign it to FingerNil.
// To re-include it, assign it to a valid finger.
//
// Note: the key with coordinates 4x, -0.58y (see images in doc) is
//       considered Row 5, Column 5 for simplicity.
//
// If you want to adjust a key's weight, change its 3rd field.
// For example, the weight of {0, 2, 1.5, FingerMiddle} is 1.5.

var KeyInfos = [NumKeys]KeyInfo{
	// Row 1
	{-2, 1.67, 2.5, FingerMiddle}, {-1, 1.76, 2, FingerMiddle}, {0, 2, 1.5, FingerMiddle}, {1, 2.08, 2, FingerIndex}, {2, 2, 2.5, FingerIndex}, {3, 1.92, 3, FingerIndex},
	// Row 2
	{-2, 0.67, 2, FingerRing}, {-1, 0.76, 1.5, FingerRing}, {0, 1, 1, FingerMiddle}, {1, 1.08, 1.5, FingerIndex}, {2, 1, 2, FingerIndex}, {3, 0.92, 2.5, FingerIndex},
	// Row 3
	{-2, -0.33, 1.5, FingerRing}, {-1, -0.24, 1, FingerRing}, {0, 0, 1.5, FingerMiddle}, {1, 0.08, 1, FingerIndex}, {2, 0, 1.5, FingerIndex}, {3, -0.08, 2.5, FingerIndex},
	// Row 4 -- key reserved for pinky excluded by default
	{-2, -1.33, 2, FingerNil}, {-1, -1.24, 1.5, FingerRing}, {0, -1, 2, FingerMiddle}, {1, -0.92, 1.5, FingerIndex}, {2, -1, 2, FingerIndex}, {3, -1.08, 2.5, FingerIndex},
	// Row 5 -- keys reserved for thumb excluded by default
	{0.5, -2, 2.5, FingerMiddle}, {1.5, -2, 2.5, FingerIndex}, {2.5, -2.08, 3, FingerNil}, {3.8, -2.09, 3.5, FingerNil}, {4, -0.58, 3.5, FingerIndex},
}

// Place your symbols you want mapped below. Please make sure the symbols
// belong to the language you chose, and that their number is the same as
// the number of keys included. Otherwise, an error will occur.

const TargetSymbols = "abcdefghijklmnopqrstuvwxyz"

// For the available languages and their alphabets, see the README.

const TargetLanguageCode = "en"
