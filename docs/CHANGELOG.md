# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Go's Versioning](https://go.dev/doc/modules/release-workflow).

## [Unreleased]
### Added
- SVS: StateVectors are now ordered by latest entries.
- SVS Scheduler now properly adds randomness to values.
- Users of SVS can now define the initial fetcher queue length.

### Changed
- SVS: Stop calling `go` on every updateCallback within Core.
- SVS Fetcher now uses a channel of functions rather than a channel of structs.

### Removed
- SVS FetchItem is removed due to not being used.

## [v0.0.0-alpha.1] - 2022-10-18
### Added
- SVS Implementation according to Specification with a built-in Fetcher
- SVS Examples: low-level (only-core, count) and high-level (count, chat)

### Security
- SVS does is not secure due to having lack signing / validating capabilities (waiting on go-ndn)

[Unreleased]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.1...HEAD
[v0.0.0-alpha.1]: https://github.com/justincpresley/ndn-sync/releases/tag/v0.0.0-alpha.1
