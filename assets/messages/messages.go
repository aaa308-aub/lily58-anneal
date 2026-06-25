package messages

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	cfg "github.com/aaa308-aub/lily58-anneal/config"
)

const numSymbols = len(cfg.TargetSymbols)

type ThreadMessage struct {
	ThreadID int
	Message  string
}

// Global channels are generally discouraged, but I prefer careful
// lifetime and ownership over having to pass it everywhere. One
// man's bad practice is another man's idiom. If you are forking
// this codebase and dislike that, I do apologize for the
// inconvenience of having to rewrite this.

var MainChannel = make(chan ThreadMessage, 100)
var keys = cfg.KeyInfos

func (msg ThreadMessage) Format() string {

	return "[SA Thread #" + strconv.Itoa(msg.ThreadID) + "] " + msg.Message
}

func PrintMessages(ch chan ThreadMessage, wg *sync.WaitGroup) {

	defer wg.Done()

	for msg := range ch {
		fmt.Print(msg.Format())
	}
}

// Requires modification if codebase is forked for a different keyboard.
func FormatLayout(layout *[numSymbols]int) string {

	// Expected output format:
	//
	// [00][01][02][03][04][05]\n
	// [06][07][08][09][10][11]\n
	// [12][13][14][15][16][17] [28]\n
	// [18][19][20][21][22][23]\n
	// \5 spaces\[24][25][26][27]
	//
	// The last key ( [28] ) is a special case because it's considered
	// row 4, column 4 (0-indexed). Moving it to the right place
	// requires use of slices.

	// Layout is a mapping of symbols to keys. Invert the bijection
	// for keys to symbols.
	var keyToSymbol [numSymbols]int
	for symbolIdx, keyIdx := range layout {
		keyToSymbol[keyIdx] = symbolIdx
	}

	// Define every row's number of keys, and how many padded spaces
	// before each row.
	rowInfos := [...]struct{ numKeys, numSpaces int }{
		{6, 0}, {6, 0}, {6, 0}, {6, 0}, {5, 7},
	}

	type rowT []string
	rows := make([]rowT, 0, 100)

	keyIdx, layoutIdx := 0, 0
	for rowIdx, rowInfo := range rowInfos {

		rows = append(rows, make(rowT, 0, 100))
		paddedSpaces := strings.Repeat(" ", rowInfo.numSpaces)
		rows[rowIdx] = append(rows[rowIdx], paddedSpaces)

		for offset := keyIdx; keyIdx < offset+rowInfo.numKeys; keyIdx++ {
			key := keys[keyIdx]

			if key.AssignedFinger == cfg.FingerNil {
				rows[rowIdx] = append(
					rows[rowIdx],
					"["+string(cfg.ExcludedKeySymbol)+"]",
				)
			} else {
				symbolIdx := keyToSymbol[layoutIdx]

				rows[rowIdx] = append(
					rows[rowIdx],
					"["+string(cfg.TargetSymbols[symbolIdx])+"]",
				)
				layoutIdx++
			}
		}

		rows[rowIdx] = append(rows[rowIdx], "\n")
	}

	// I expanded the instructions to make them a bit more readable.
	numRows := len(rows)
	oldRow := rows[numRows-1]
	keyPosToMove := len(oldRow) - 2 // Before the newline.
	specialKeyStr := oldRow[keyPosToMove]
	oldRow = slices.Delete(oldRow, keyPosToMove, keyPosToMove+2) // Remove newline too.
	specialKeyStr = " " + specialKeyStr
	rows[2] = slices.Insert(rows[2], len(rows[2])-1, specialKeyStr) // Before the newline.

	// There's less than 50 strings to join, so keep it simple.
	str := ""
	for _, row := range rows {
		str += strings.Join(row, "")
	}

	return str
}

func FormatFinalMessage(
	score float32,
	layout *[numSymbols]int,
) string {
	return fmt.Sprintf("Finished SA run with final score of %f\n", score) +
		FormatLayout(layout) + "\n\n"
}
