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

This will make the Simulated Annealing (SA) algorithm naturally gravitate towards the center. The center is technically the key on the 3rd row, 3rd column, but pulled slightly upward to align with the fingers' natural "resting" position.\
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
Note that all the other variables you see within ``config.go`` are not meant to be changed, but of course do whatever you want (if you know what you're doing!).

## Supported Languages & Corresponding Symbols
The engine currently supports 39 different languages with their corresponding symbols. Feel free to open an issue or to contact me if there's any confusion.

| Language | Code | Symbols |
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
