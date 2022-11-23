# Changes Made

All substantial changes/modifications/additions away from [the original repository](https://github.com/elliotchance/orderedmap) are documented in this file.
These changes were performed on `v2` of the original repository at the commit `1e43e194ff533a346bab5f9b66b738256f199c8a`.

## Added
- list has the functions MoveToBack() and MoveToFront() to move Elements.

## Changed
- list now uses a head and tail pointer rather than a root node.
- OrderedMap's Set() now take in whether the value is old or new. This ideally keeps an order of newness. If the value is new, add or change AND push to front. If the value is old, simply add to back OR just change.
- Various other small changes.

## Removed
- OrderedMap's Keys() and GetOrDefault() functions.
