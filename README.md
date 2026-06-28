# LILY58-ANNEAL
## Introduction
This is an optimization engine to find a near-optimal layout for a left-hand only Lily58 keyboard, using a Simulated Annealing (SA) algorithm that I wrote, prioritizing flow and hand comfort.

Lily58 is a split keyboard that boasts 29 keys on each side (or 28 if you replace a certain key with a dial/knob). This engine cannot find you the best layout for the entire split; you can find other engines on the internet for that. Its goal is to strictly find the best layout on one side of the Lily58 by mapping the desired characters to the desired keys, using data I sourced for 39 different languages.

## How to Install & Run
1. You need [Go installed](https://go.dev/doc/install) to compile this program, and it must be version ``1.22`` or above. There are no other dependencies or go packages required.
2. Open a terminal in your desired folder and clone the repository (which will create a parent folder inside named ``lily58-anneal``):
```
git clone https://github.com/aaa308-aub/lily58-anneal
```
3. Build and run:
```
cd lily58-anneal
go build .
./lily58-anneal
```
4. Wait for the program to print the engine's results. Should take less than a minute on an average, modern CPU. It will print messages in this form:
```
[SA Thread #4] Started SA run with initial score of 6.903992
[SA Thread #2] Started SA run with initial score of 6.700186
...

[SA Thread #6] Finished SA run with final score of 1.866792
[v][w][k][b][z][q]
[d][t][h][a][c][g]
[l][n][o][e][i][j] [x]
[·][r][u][s][y][f]
       [m][p][·][·]

[SA Thread #1] Finished SA run with final score of 1.847535
[k][l][p][f][y][z]
[m][r][s][o][u][q]
[c][t][h][e][i][w] [x]
[·][d][n][a][g][j]
       [b][v][·][·]
...

```
5. Assess the output. Note that the lower score, the better the layout (usually).

## Configuration & Limitations
All configuration happens in the config file ``lily58-anneal/config/config.go``. This short [documentation](doc.md) lets you know what you can change. ``TL;DR:`` You can include or ignore any of the 29 keys available and assign which finger presses each key. However you cannot assign the thumb or pinky to any keys (reasons are in the documentation). You must choose a supported language and the symbols (letters) to map. Those symbols must belong to the language's alphabet (see section below).

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

## Special Thanks
Getting only one half of the Lily58 keyboard without purchasing the whole split can be tricky, but [Typeractive](https://typeractive.xyz/) made it possible and I thank them dearly for their kindness and support for this project.
