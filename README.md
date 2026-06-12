# LILY58-ANNEAL
## Introduction
This is an optimization engine to find the best layout for a left-hand only Lily58 keyboard, using a Simulated Annealing algorithm that I wrote, prioritizing flow and hand comfort.

Lily58 is a split keyboard that boasts 29 keys on each side (or 28 if you replace a certain key with a dial/knob). This engine cannot find you the best layout for the entire split; you can find other engines on the internet for that. Its goal is to strictly find the best layout on one side of the Lily58, given a certain number of keys and the characters to map them to, and fed your own data from typing (code, essays, text messages, etc.).

## Implementation
By default, the engine finds the best layout for my own left-hand only Lily58 to map 26 keys to the English alphabet. I have restricted the thumb to two keys only for Space and Enter and the pinky to one key only for Backspace. This means that **the engine assigns only the Ring, Middle and Index fingers to these 26 keys, and while you can choose another set of keys and which finger presses what key, you cannot assign the pinky or thumb to any key -- the engine doesn't know about them.** This was a deliberate design choice that I made, knowing that the thumb has different musculature and runs on a small and distant set of keys and that the pinky is a weak finger with little travel. The engine is responsible for the other 3 fingers and to avoid stretching or pressing two keys in a row with the same finger (called Same-Finger Bigram).

If you'd like to use this engine but for different key mappings, [take a look at this short documentation](doc.md).
