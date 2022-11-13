# Changes Made

All substantial changes/modifications/additions away from [the original repository](https://github.com/reddragon/bloomfilter.go) are documented in this file.
These changes were performed on the original repository at the commit `5c5f79c7342e6d65c9356aecf7c8b4cf30089809`.

## Added
- Byte Functionality (Encoding + Decoding)

## Changed
- Utilize xxhash3-64 for the hash function instead of fnv.
- Use an array of bytes (uint8) rather than an array of bools for the bitmap.
- Change variables to uint rather than int.
- Make NewBloomFilter() more readable.
- Rename functions and structs.
- Ensure k hash functions.

## Removed
- All CountingBloomFilter functionality.
- All ScalableBloomFilter functionality.
- Counting elements in the Filter.