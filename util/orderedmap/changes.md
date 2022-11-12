# Changes Made

Repository: github.com/wk8/go-ordered-map

All substantial changes/modifications/additions away from the original repository are documented in this file.
These changes were performed on the original repository at the commit `ab40b2b9a64414289e84274611d1fee69e52ff55`.

## Added
- New External API Functions: Clone() / Copy(), Reverse(), Size(), First() / Front(), Last() / Back() which are self-described.
- New Internal Function setNewPair() to prevent code copying.

## Changed
- Set() / Load() now take in whether the value is old or new. This ideally keeps an order of newness. If the value is new, add or change AND push to front. If the value is old, simply add to back OR just change.
- Testing now covers the new functions that were added.
