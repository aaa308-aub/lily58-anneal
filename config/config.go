package config

type finger uint8

const (
	FingerNil    finger = iota // 0
	FingerRing                 // 1
	FingerMiddle               // 2
	FingerIndex                // 3
)

type keyInfo struct {
	X, Y, Weight   float32
	AssignedFinger finger
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
//
// Adjust the number of keys included below when you're done.
// Please make sure it's correct, or an error will occur.

const NumKeysIncluded = 26

var KeyConfig = [...]keyInfo{
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
