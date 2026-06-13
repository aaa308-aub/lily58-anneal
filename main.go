package main

import (
	"fmt"
	"unicode"

	"github.com/aaa308-aub/lily58-anneal/parser"
)

type finger int

const (
	invalidFinger finger = iota // 0
	ring                        // 1
	middle                      // 2
	index                       // 3
)

type keyInfo struct {
	X, Y, Weight   float64
	AssignedFinger finger
}

func main() {

	keyInfos := [...]keyInfo{
		// Row 1
		{-2, 1.67, 2.5, middle}, {-1, 1.76, 2, middle}, {0, 2, 1.5, middle}, {1, 2.08, 2, index}, {2, 2, 2.5, index}, {3, 1.92, 3, index},
		// Row 2
		{-2, 0.67, 2, ring}, {-1, 0.76, 1.5, ring}, {0, 1, 1, middle}, {1, 1.08, 1.5, index}, {2, 1, 2, index}, {3, 0.92, 2.5, index},
		// Row 3
		{-2, -0.33, 1.5, ring}, {-1, -0.24, 1, ring}, {0, 0, 1.5, middle}, {1, 0.08, 1, index}, {2, 0, 1.5, index}, {3, -0.08, 2.5, index},
		// Row 4 -- key reserved for pinky excluded by default
		/*{-2, -1.33, 2, invalidFinger},*/ {-1, -1.24, 1.5, ring}, {0, -1, 2, middle}, {1, -0.92, 1.5, index}, {2, -1, 2, index}, {3, -1.08, 2.5, index},
		// Row 5 -- keys reserved for thumb excluded by default
		{0.5, -2, 2.5, middle}, {1.5, -2, 2.5, index} /*{2.5, -2.08, 3, invalidFinger}, {3.8, -2.09, 3.5, invalidFinger},*/, {4, -0.58, 3.5, index},
	}

	targetSymbols := []rune{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	}

	// Note: the key with coordinates 4x, -0.58y is considered Row 5, Column 5 for simplicity.
	//
	// How to exclude/re-include keys:
	// To exclude a key, simply block-comment it out as seen above. To re-include it, remove the block comment and assign a valid finger.
	//
	// You can add or remove any symbols to your liking, including hidden ones like Enter ( '\n' ) or Tab ( '\t' ).
	//
	// If you want to adjust a key's weight, change its 3rd field. For example, the weight of {0, 2, 1.5, middle} is 1.5.
	//
	// Make sure the number of target symbols matches the number of included keys.

	numKeys, numSymbols := len(keyInfos), len(targetSymbols)

	if numKeys > 29 || numKeys < 2 {
		panic(fmt.Sprintf(
			"Config Error: Number of keys (%d) must be between 2 and 29",
			len(keyInfos),
		))
	}

	if numKeys != numSymbols {
		panic(fmt.Sprintf(
			"Config Error: Number of symbols (%d) does not match number of keys (%d)",
			numKeys,
			numSymbols,
		))
	}

	for i := range keyInfos {
		if keyInfos[i].AssignedFinger == invalidFinger {
			panic(
				"Config Error: Key assigned to invalidFinger, most likely forgot to assign properly",
			)
		}
	}

	for i := range targetSymbols {
		targetSymbols[i] = unicode.ToLower(targetSymbols[i])
	}

	matrix, err := parser.FillBigramMatrix("./data.txt", targetSymbols)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", matrix) // temporary
}
