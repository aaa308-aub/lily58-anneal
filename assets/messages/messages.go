package messages

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	cfg "github.com/aaa308-aub/lily58-anneal/config"
)

const nSym = cfg.NumSymbols

type ThreadMessageT struct {
	ID  int
	Msg string
}

func (m ThreadMessageT) Format() string {

	return "[SA Thread #" + strconv.Itoa(m.ID) + "] " + m.Msg
}

func PrintMessages(ch chan ThreadMessageT, wg *sync.WaitGroup) {

	defer wg.Done()

	for m := range ch {
		fmt.Print(m.Format())
	}
}

// Requires modification if codebase is forked for a different keyboard.
func FormatLayout(layout *[nSym]int) string {

	// Expected output format:
	//
	// [00][01][02][03][04][05]\n
	// [06][07][08][09][10][11]\n
	// [12][13][14][15][16][17] [28]\n
	// [18][19][20][21][22][23]\n
	// \7 spaces\[24][25][26][27]
	//
	// The last key ( [28] ) is a special case because it's considered
	// row 4, column 4 (0-indexed). Moving it to the desired place
	// requires use of slices.

	var keys, syms = cfg.KeysAll, cfg.SymbolsArr
	const nilSym = cfg.ExcludedKeySymbol

	// Layout is a mapping of symbols to keys. Must invert.
	var keyToSym [nSym]int
	for symIdx, keyIdx := range layout {
		keyToSym[keyIdx] = symIdx
	}

	// Define each row's number of keys, and how many
	// leading spaces before each row.
	rowParams := [...]struct{ nKey, nSpace int }{
		{6, 0}, {6, 0}, {6, 0}, {6, 0}, {5, 7},
	}

	type rowT []string
	rows := make([]rowT, 0, 100)

	keyIdx, layoutIdx := 0, 0
	for rowIdx, p := range rowParams {
		nK, nS := p.nKey, p.nSpace

		rows = append(rows, make(rowT, 0, 100))
		spaces := strings.Repeat(" ", nS)
		rows[rowIdx] = append(rows[rowIdx], spaces)

		for offset := keyIdx; keyIdx < offset+nK; keyIdx++ {
			key := keys[keyIdx]

			if key.Fin == cfg.FingerNil {
				rows[rowIdx] = append(
					rows[rowIdx],
					"["+string(nilSym)+"]",
				)
			} else {
				symbolIdx := keyToSym[layoutIdx]

				rows[rowIdx] = append(
					rows[rowIdx],
					"["+string(syms[symbolIdx])+"]",
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

	// The number of strings to join is tiny so this is fine.
	str := ""
	for _, row := range rows {
		str += strings.Join(row, "")
	}

	return str
}

// Does not format or include initial layout because it's irrelevant.
func FormatInitialMessage(
	score float32,
) string {
	return fmt.Sprintf("Started SA run with initial score of %f\n", score)
}

func FormatFinalMessage(
	score float32,
	layout *[nSym]int,
) string {
	return fmt.Sprintf("Finished SA run with final score of %f\n", score) +
		FormatLayout(layout) + "\n\n"
}
