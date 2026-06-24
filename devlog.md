This is a technical diary or footprint that describes every ``Problem`` faced\
along with its ``Constraints`` and the applied ``Solution``, in order of\
procedure of development (oldest ``Problem`` to newest). It follows the format:

```
#<problem number>
<problem description>
<empty line>
<constraints description>
<empty line>
<solution description>
<empty line>
<empty line>
#<next problem number>
...
```

#1\
The right hand is heavily taxed with both typing on the keyboard and navigating\
with the mouse, leading to excessive hand movement between the two and\
disrupting workflow.

Replacing the mouse with shortcuts like in Vim/NeoVim is not an acceptable\
solution, because it undermines the mouse as a screen navigation tool.

Use a left-hand only keyboard layout to replace a regular keyboard, sacrificing\
speed for flow. Choose the Lily58 for its ergonomic design and ideal key count\
of 29. It also supports key-map layering to accomodate more than 29 inputs.


#2\
There is no one-handed Lily58 keyboard layout to be found online. A custom\
layout must be made.

Must fit 26 inputs for the English alphabet in a dense 29-key layout. Must also\
include Space, Backspace and Enter, leaving no room for other inputs in the\
"main layer". Must optimize for hand/finger efficiency and minimize discomfort\
from stretching or same-finger bigrams (e.g., ``ring->ring->middle->middle...``)

Design and implement a keyboard layout optimization engine using a simulated\
annealing algorithm. Reserve the thumb to Space and Enter and the weak pinky to\
Backspace. This simplifies the engine's purpose to optimize for just 3 fingers.\
Use fixed weights as cost/penalty that increases with distance from keyboard\
center, and scale with monogram/symbol frequency. Make measurements on keyboard\
to penalize excess stretching between fingers scaled by distance and bigram\
frequency. Use trigram data to reward (apply negative cost to) inward rolling\
with fingers (``ring->middle->index``).


#3\
Someone who's interested in their own one-handed Lily58 may use this engine,\
but for a language different from English.

``No notable constraints.``

Engine supports any inputs/symbols. Written in Go to standardize symbols with\
runes encoded in ``UTF-8``.


#4\
N-gram data is required for solution proposed in ``#2.``

Data must be from large sample sizes, especially for trigrams, because there\
are 26^3 = 17576 possible trigram combinations for English, and much more for\
some languages.

Initially allowed users to provide their own corpus to be parsed for n-gram\
frequencies, but millions of characters would be required to accurately\
present the top trigrams and their frequencies. Shifted focus to sourcing\
n-gram statistics from various archives for about 40 languages. Source for\
English is ``norvig.com/mayzner.html``. Wrote short scripts to standardize data\
in .tsv files in this format:
```
NGRAM1\tCOUNT\n
NGRAM2\tCOUNT\n
...
```
Organized vetted data in this file structure:
```
 assets/
└── counts/
    ├── monograms/
    │   ├── en.tsv
    │   ├── fr.tsv
    │   └── ...
    │
    ├── bigrams/
    │   ├── en.tsv
    │   ├── fr.tsv
    │   └── ...
    │
    └── trigrams/
        ├── en.tsv
        ├── fr.tsv
        └── ...
```
Note that the engine works with frequencies (counts over the total), not\
counts. I decided to keep the data in terms of counts because it's more\
informative, in case anyone wants to use these files themselves. The\
initialization cost of turning them to frequencies is negligible.


#5\
Southeast Asian countries like Japan, Korea and China practice typing very\
different from other countries, like joining simple letters to form new ones,\
leading to complications in n-gram data and what it may represent.

``No notable constraints.``

Removed n-gram data for Southeast Asian languages due to high language-specific\
complexity that cannot be handled in this project, at least for the moment.


#6\
Users should be able to easily change key configurations, the letters/symbols\
to be mapped, and the target language.

Configurations should be done at compile-time to bring data to the engine in\
stack-allocated data structures and maximize performance.

Wrote config.go for easy configuration of key info (XY-coords, weights, finger\
assignments), target language (``en``, ``fr``, ``de``, etc.) and target\
symbols, using my configs as defaults. Number of keys/symbols can be found as\
constants to bring n-gram data in arrays or flattened matrices rather than\
slices, preventing heap allocations and pointer chasing once the actual engine\
starts. Simulated annealing is an algorithm which takes millions of randomized\
steps per run and thus my data structures must be data-oriented and mostly\
stack-allocated.


#7\
Bigram data must be stored in such a way that is data-oriented: the best\
and fastest way the engine can access it.

The biggest possible number of target symbols is 29. Bigrams are ordered pairs\
with repetition, so there are only 29^2 = 841 bigram frequencies to store in\
the worst case. This is obviously a relief rather than a constraint.

Implemented function GetBigramData which fills a flattened 2D matrix of bigram\
frequencies by cross-tabulating the string of target symbols with itself. For\
example, if the symbols are ``abc``, the frequencies are stored like this:
```
[aa][ab][ac]
[ba][bb][bc]
[ca][cb][cc]
```
The simulated annealing algorithm swaps 2 random keys/symbols every step. If\
they are ``a`` and ``c`` in this example, it will calculate the cost delta of\
bigrams containing ``a`` and ``c`` only.


#8\
Trigram data must be stored and accessed efficiently.

There are over 17500 possible trigram combinations for 26 or more symbols.\
Tracking every trigram would be wasteful. To reward inward rolling with fingers\
on trigrams, the order of symbols in a trigram must be known.

Analyzed the data and saw that roughly 33.6% of English trigrams, in terms of\
cumulative frequency, are concentrated in the top 100 trigrams. Fixed number of\
top trigrams stored in data to 100 under the assumption that most languages\
follow the same trend, and that more trigrams is too much noise to optimize for\
any realistic one-handed keyboard layout. Implemented these data structures:
```Go
type TrigramInfo struct {
	Freq           float32
	orderedSymbols [3]int8
}

var trigramInfos [numTopTrigrams]trigramInfo // where numTopTrigrams = 100

var symbolToTrigrams [numSymbols*numTopTrigrams]int8
```
``symbolToTrigrams`` is a flattened 2D matrix which takes the index of each\
symbol and returns a bucket of indices to the trigrams it belongs to, starting\
at offset ``numTopTrigrams * symbolIndex`` with a size of ``numTopTrigrams``.\
This maximizes lookups but wastes some memory because the matrix is sparse.\
Rejected turning it into a Compressed Sparse Matrix (CSF) because it would lead\
to pointer chasing and/or\ more expensive indexing.


#9\
The engine uses bigram frequency data to punish same-finger bigrams and/or\
stretches, regardless which key is pressed first. So the data should be for\
unordered bigrams by aggregating ordered data (e.g.,``"ab" + "ba"``). The data\
structure proposed in ``#7`` is inefficient because it stores data for ordered\
bigrams, and on every swap, the engine has to make twice the number of passes\
needed, were the data for unordered bigrams.

Bigrams with non-distinct symbols, like ``aa``, ``bb``... should be ignored,\
because the engine can't optimize for them.

Assessed multiple data structures to solve this problem, but the only solution\
that is readable, cheap to index, and constraint-free is a symmetric matrix\
with a zero-value diagonal. For example, for symbols ``abcd``, we would have\
this new matrix of bigram frequencies, but flattened:
```
    OldTable              NewTable
[aa][ab][ac][ad]      [00][ab][ac][ad]
[ba][bb][bc][bd]      [ab][00][bc][bd]
[ca][cb][cc][cd]      [ac][bc][00][cd]
[da][db][dc][dd]      [ad][bd][cd][00]
...Where NewTable[i,j] = OldTable[i,j] + OldTable[j,i]
```
With the new table, if the engine swaps the symbols ``b`` and ``d``, we simply\
re-evaluate the 2nd and 4th row -- 8 entries. With the old table, we would\
have to re-evaluate the "cross" formed by each symbol -- 14 entries -- and\
also deal with double-counting of bigrams containing both swapped symbols.\
Although time is wasted processing the zero-entries, it is negligible and\
arguably necessary for readability and maintenance.

#10\
The solution proposed in ``#8`` wastes a lot of memory and introduces an issue\
with double-counting trigrams which happen to contain both swapped symbols.\

``No notable constraints.``

Replaced wasteful sparse matrix with a bitmask for each symbol, where each bit\
represents whether the symbol belongs to the trigram in ``trigramInfos``, with\
the index equal to that bit's position. Example:
```Go
var symbolToTrigrams [numSymbols]uint64

symbolToTrigrams[5] = 0b1010
// This means that the symbol indexed 5 belongs to trigrams 1 and 3 in
// trigramInfos. To extract the indices:
for x = symbolToTrigrams[5]; x != 0; x &= (x-1) {
	i := bits.TrailingZeros64(x)
	// Process trigramInfos[i]
}
```
Prevented double-counting of trigrams that contain both swapped symbols during\
our engine, by simply taking the union of their ``symbolToTrigram`` entries\
(bitwise ``OR``). But we must drop the number of trigrams tracked from 100 to\
64 to use this solution. This might be better for the engine to prioritize the\
trigrams that matter the most, anyway.

#11\
``WholeCostLayout`` of output layout does not match the cost (score) that is\
being calculated by adding ``deltaCost`` every swap. There is a logical error\
somewhere.

``No notable constraints.``

Implemented ``bigramCost`` function mainly to un-count the double-counted\
bigram on every swap that happens to contain both swapped symbols, which was\
naively thought to be implicitly handled by the data structure implemented for\
problem ``#9``. ``WholeCostLayout`` is now almost equal to the output score,\
with slight difference due to float32 rounding errors. I simply stopped adding\
``deltaCost`` to the score every swap (when it's accepted) because there's just\
no need, and instead use ``WholeCostLayout``'s output to print the final score.
