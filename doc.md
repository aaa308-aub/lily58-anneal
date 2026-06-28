## Default Configuration & Finger Assignment
By default, the engine finds the best layout for my own left-hand only Lily58 to map 26 keys to the English alphabet. I have restricted the thumb to two keys only for Space and Enter and the pinky to one key only for Backspace. This means that the engine assigns only the Ring, Middle and Index fingers to these 26 keys, and while you can choose another set of keys and which finger presses what key, you cannot assign the pinky or thumb to any key -- the engine doesn't know about them. Here are the default finger assignments:

![](assets/images/lily58fingers.png)

You'll notice almost immediately that the index finger is responsible for half the keyboard: 14 keys. My goal with the left-hand only Lily58 has always been to keep my right hand on the mouse and stop moving it to the keyboard and back. So, my priority was always flow over speed, and that is why I accept the speed bottleneck that will obviously come with this index tax.

## Key Distances from the Center
These are the measurements I've made on an actual Lily58:

![](assets/images/lily58distances.png)

The unit of measurement is the distance between the centers of two adjacent keys. The Lily58 follows the standard unit for keyboards ``1u = 19.05mm``.\
These measurements are used to punish stretches, like stretching the ring finger to the left then the index to the right, which creates an uncomfortable hand position.

## Gravitation towards the Center
A flat penalty (or weight) is applied to every key pressed, and that penalty increases the further away it is from the center of the keyboard:

![](assets/images/lily58weights.png)

This will make the Simulated Annealing (SA) algorithm naturally gravitate towards the center. The center is technically the key on the row-column (3,3) as seen in the section above. But for these weights, I pulled the center slightly upward to align with the fingers' natural "resting" position.\
You may wonder why I didn't just use the distances I've measured for this. It's because these distances can't capture the "feel" of your hand on each key, and I decoupled the logic of weights from distance for you to be able to easily change them to your liking.

## How to Configure to your Needs
In ``config.go``, within ``lily58-anneal/config/`` folder, the variables you can change are listed below. The comments inside the file provide more detail.
```Go
// Key information is in this format:
// {X-coordinate, Y-coordinate, Weight, Assigned Finger}
var KeysAll = [NumKeysAll]KeyT{
	// Row 0
	{-2, 1.67, 2.5, FingerMiddle}, // ...
	// Row 1
	{-2, 0.67, 2, FingerRing}, // ...
	// Row 2
	{-2, -0.33, 1.5, FingerRing}, // ...
	// Row 3 -- FingerNil means the key is excluded and won't be seen by the engine.
	{-2, -1.33, 2, FingerNil}, // ...
	// Row 4
	{0.5, -2, 2.5, FingerMiddle}, // ...
}

// Place the symbols you want mapped below, making sure the symbols
// belong to the language you chose and that their number is the
// same as the number of keys included.

const symbolsStr = "abcdefghijklmnopqrstuvwxyz"

// For the available languages and their alphabets, see the section
// below. The same table is copied in the README.

const TargetLanguageCode = "en"
```
Note that all the other variables you see within ``config.go`` are not meant to be changed, but of course, do whatever you want if you know what you're doing.

## Supported Languages & Corresponding Symbols
The engine currently supports 39 different languages with their corresponding symbols. Feel free to open an issue or to contact me if there's any confusion.

| Language | Code | Alphabet |
|-|-|-|
| Afrikaans | af | abcdefghijklmnopqrstuvwxyzàáâèéêëîïôóúûü |
| Bulgarian | bg | абвгдежзийклмнопрстуфхцчшщъьюя |
| Bosnian | bs | abcdefghijklmnoprstuvzžćčđš |
| Catalan | ca | abcdefghijklmnopqrstuvwxyzàçéèíïóòúü |
| Czech | cs | abcdefghijklmnopqrstuvwxyzáčďéěíňóřšťúůýž |
| Danish | da | abcdefghijklmnopqrstuvwxyzæøå |
| German | de | abcdefghijklmnopqrstuvwxyzäëöß |
| Greek | el | αβγδεζηθικλμνξοπρστυφχψως |
| English | en | abcdefghijklmnopqrstuvwxyz |
| Esperanto | eo | abcdefghijklmnoprs-tuŭvzĉĝĥĵŝ |
| Spanish | es | abcdefghijklmnopqrstuvwxyzáéíóúüñ |
| Estonian | et | abdefghijklmnoprstuvzžšõäöü |
| Finnish | fi | abcdefghijklmnopqrstuvwxyzåäö |
| French | fr | abcdefghijklmnopqrstuvwxyzàâæçéèêëîïôœùûüÿ |
| Galician | gl | abcdefghilmnñopqrstuvxz |
| Croatian | hr | abcdefghijklmnoprstuvzžšđčć |
| Hungarian | hu | abcdefghijklmnoprstuvzáéíóöőúüű |
| Indonesian | id | abcdefghijklmnopqrstuvwxyz |
| Icelandic | is | aábdðeéfghiíjklmnoóprstuúvxyýþæö |
| Italian | it | abcdefghilmnopqrstuvzáèéìíòóùú |
| Georgian | ka | აბგდევზთიკლმნოპჟრსტუფქღყშჩცძწჭხჯჰ |
| Lithuanian | lt | abcdefghijklmnoprstuvząčęėįšųūž |
| Latvian | lv | abcdefghijgklmnoprstuvzāčēģīķļņšūž |
| Macedonian | mk | абвгдѓежзѕијклљмнњопрстќуфхцчџш |
| Malay | ms | abcdefghijklmnopqrstuvwxyz |
| Dutch | nl | abcdefghijklmnopqrstuvwxyzáéíóúàèëïöü |
| Norwegian | no | abcdefghijklmnopqrstuvwxyzæøå |
| Polish | pl | aąbcćdeęfghijklłmnńoóprsśtuwyzźż |
| Romanian | ro | abcdefghijlmnoprstuvxzăâîșțşţ |
| Russian | ru | абвгдеёжзийклмнопрстуфхцчшщъыьэюя |
| Slovak | sk | abcdefghijklmnoprstuvxyzáäčďéíĺľňóôŕšťúýž |
| Slovenian | sl | abcdefghijklmnoprstuvzčšž |
| Albanian | sq | abcdefghijklmnopqrtuvxyzçë |
| Serbian | sr | абвгдђежзијклљмнњопрстћуфхцчџш |
| Swedish | sv | abcdefghijklmnopqrstuvwxyzåäö |
| Thai | th | กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรลวศษสหฬอฮะัาำิีึืุูเแโใไๅ็่้๊๋์ํู |
| Turkish | tr | abcçdefgğhıijklmnoöprsştuüvyz |
| Ukrainian | uk | абвгґдеєжзиіїйคลмнопрстуфхцчшщьюя |
| Vietnamese | vi | abcdefghijklmnopqrtuvwxyzaăâeêioôơuưyáàảãạắằẳẵặấầẩẫậéèẻẽẹếềểễệíìỉĩịóòỏõọốồổỗộớờởỡợúùủũụứừửữựýỳỷỹỵđ |

## How the Engine Works
I want to keep the doc lean, so I'll only capture the ideas and not the math here.

[Simulated Annealing](https://en.wikipedia.org/wiki/Simulated_annealing), in simple terms, is an algorithm that randomly walks a search space and tries to find its lowest or highest point -- in our case it's lowest. It has no intuition what the next step -- which is swapping two symbols by key -- should be, as it's completely random. Whenever it takes a step that results in a lower point/score, the new state resulting from that step is always accepted as the current state. If the step results in a higher (a worse) score, instead of rejecting the new state, a probability of acceptance is calculated. That *chance* is inversely proportional to the time elapsed since the algorithm began its run, and to how bad the difference of scores -- the new score minus the old score -- would be if the new state is accepted.

In the ``assets/counts/`` folder, you'll see counts of n-grams from highest to lowest, for ``n = 1, 2, 3``. An n-gram is an ordered combination of ``n`` symbols.\
Monograms are used to penalize layouts with high-frequency symbols, like ``e`` or ``t`` in English, placed far away from the keyboard center. Bigrams like ``he`` or ``in`` are for punishing keys pressed consecutively with the same finger or with excessive finger stretching (or both). Lastly, layouts where distinct fingers (like ``ring->middle->index`` or ``middle->ring->index``) are used to press some of the most frequent trigrams, like ``the`` or ``and`` in English, are rewarded with lower scores.

There are other properties you *could* optimize for such as excessive alternation between the top and bottom rows, or to target more trigrams than what I've set (the top 64 trigrams), but the choices become more limited with a one-hand keyboard, since each property accounted for would make the search space more "rugged" than it already is. A more rugged space leads to generally worse results.
