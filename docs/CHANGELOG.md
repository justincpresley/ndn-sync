# Changelog

All notable changes to syncs will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Go's Versioning](https://go.dev/doc/modules/release-workflow).

## [Unreleased]

## [v0.0.0-alpha.4] - 2022-11-13
- All Syncs in SVS now implement retries!
- Standardize the seqno within SVS to a uint64.
- Utilize go-ndn's methods for encoding.
- Exposed all internal through util due to necessary access.
- Modified to ensure compatibility to go-ndn's latest changes.
- BloomFilter code. (for future plans)

## [v0.0.0-alpha.3] - 2022-10-26
### Changed
- SVS: Pulled out init() into its own file.
- Utilize TLNum (instead of uint) for SVS TlvTypes.
- Fixed StateVector Encoding to met specification.

## [v0.0.0-alpha.2] - 2022-10-22
### Added
- SharedSync is now available in SVS.
- SVS: StateVectors are now ordered by latest entries. (for future plans)
- SVS Scheduler now properly adds randomness to values.
- Users of SVS can now define the initial fetcher queue length.

### Changed
- SVS: Stop calling `go` on every updateCallback within Core.
- SVS Fetcher now uses a channel of functions rather than a channel of structs.

## [v0.0.0-alpha.1] - 2022-10-18
### Added
- SVS Implementation according to Specification with a built-in Fetcher
- SVS Examples: low-level (only-core, count) and high-level (count, chat)

### Security
- SVS does is not secure due to having lack signing / validating capabilities (waiting on go-ndn)

[Unreleased]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.4...HEAD
[v0.0.0-alpha.4]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.3...v0.0.0-alpha.4
[v0.0.0-alpha.3]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.2...v0.0.0-alpha.3
[v0.0.0-alpha.2]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.1...v0.0.0-alpha.2
[v0.0.0-alpha.1]: https://github.com/justincpresley/ndn-sync/releases/tag/v0.0.0-alpha.1
