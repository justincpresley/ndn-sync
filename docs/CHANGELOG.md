# Changelog

All notable changes to syncs will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Go's Versioning](https://go.dev/doc/modules/release-workflow). Moreover, ***ndn-sync*** utilizes 3 version identifiers: `alpha`, `beta`, and `mark`.

## [Unreleased]
## Added
- `EfficientSuppression` option for SVS `TwoStateCore`. Found by **@seijiotsu**, this option ignores out-of-date Sync Interests within the network RTT which dramatically reduces the number of suppressions. With extremely sparse SVS networks, this option might incorporate delay. More on this is documented [here](https://github.com/named-data/ndn-svs/issues/25) and will later be added to the Spec.

## Changed
- Slight refactor of SVS `Core`. Removed many small inefficiencies in its logic. Operations were found unnecessary in both `Suppression` and `Steady` states.
- Internal naming of variables and functions have been changed for clarity.
- Reuse of a variable during `StateVector` encoding.

## Fixed
- Slight refactor of SVS `Scheduler`, found that it was incorporating jitter wrong.
- SVS `CoreConfig` giving the appropriate `Core` type.

## [v0.0.0-alpha.13] - 2024-02-09
## Added
- `OneStateCore` option for testing purposes.

## Changed
- An SVS `StateVector` now resides within the application parameters' portion of a Sync Interest. While this is not follow the Spec currently, it will in due time as most libraries are incorporating this.
- SVS `Core` now is entirely subscription-based and can support multiple channel listeners.
- SVS `Core` is not tied to a particular dataset. You can now publish multiple datasets per node.
- Changed SVS examples to follow NFD's new default Unix socket path.
- Internal naming of variables and functions have been changed for clarity.
- Updated dependencies.

## [v0.0.0-alpha.12] - 2023-08-31
## Added
- SVS `Constants` now contain `enc.Component`s that are added in SVS's naming. This was not exposed previously.
- `BareSourceOrientedNaming` which is a new `NamingScheme` that uses no additional `enc.Component`s during SVS's naming.
- `OrderedMap`s now take an `Ordering`: `Canonical` or `LatestEntriesFirst`.
- `OrderedMap` is now less generic and more tied to our use-case of NDN. An `Element` now stores the key in both `enc.Name` and `string` forms. This results in slightly more memory usage but increases performance by minimizing the amount of 'name to string' and 'string to name' conversions throughout SVS.

## Changed
- SVS API is now `enc.Name`-based instead of being `string`-based.
- `OrderedMap` API to reflect listed changes.
- Updated dependencies.

## [v0.0.0-alpha.11] - 2023-02-03
## Added
- A new Optimized `StateVector` Encoding! Reduces 2+ bytes per entry. Set `FormalEncoding` to `false` to activate it.

## Changed
- Misspelling within SVS `Constants`.
- Updated dependencies.

## Fixed
- Data race within `Scheduler` with the pairing of `startTime` and `cycleTime`.
- Data race when resetting `heart`s within the `Tracker` of `HealthSync`.

## Removed
- Scheduler's `Add()` due to no-use. However, it can still be achieved via `Set( someTime + TimeLeft() )`.

## [v0.0.0-alpha.10] - 2023-01-06
## Added
- A new SVS Sync type `HealthSync`, an ephemeral sync for source health. It is still just a prototype however and STC.
- A new SVS example to show off `HealthSync`.

## Changed
- `Constants` now holds `time.Duration` variables.  Massively simplifies many areas of the SVS code including but not limited to `Core` and `Scheduler`.

## [v0.0.0-alpha.9] - 2023-01-01
## Added
- A new SVS `HandlingOption`! `EqualTrafficHandling` which spreads requests equally among the nodes. Please note that each handling option does have unique pros and cons.

## Changed
- SVS `NewCore()` is now a generic function taking a general `CoreConfig`. Opens the door for future options.
- Be able to alter the SVS `MissingData` structure to help track the data you still need when looping.
- SVS `Core` now provides a `chan []MissingData` rather than `chan *[]MissingData`.
- Go-ify all getters.

## [v0.0.0-alpha.8] - 2022-12-29
## Changed
- SVS `Core` now provides a missing channel instead of taking a missing data callback. More low-level control and efficiency were primary factors for this change as well as the listed fix.
- SVS `Sync`s now provide a `HandlingOption` for how the missing channel will be handled. Opens the door for future options.
- License switch to ISC. The restrictions were not very friendly.
- Renaming of variables, functions, and types.
- Other small changes.

## Fixed
- An out-of-sync `Core` vulnerability caused by having a very slow missing data callback. The results could range from just receiving updates late to missing data entirely.

## [v0.0.0-alpha.7] - 2022-12-27
## Changed
- Completely refactored SVS `Scheduler`.
- SVS's built-in fetchers for both `NativeSync` + `SharedSync` now use a channel of structs rather than a channel of funcs for readability and possible performance.
- Utilize `strings.Builder` for the `(stateVector).String()` method.
- Updated all dependencies.
- Other small changes.

## Removed
- Dependencies and code not related to the first sync, SVS.

## [v0.0.0-alpha.6] - 2022-11-24
### Changed
- `sync/atomic` is a thing and its more performant. Utilize it for `CoreState` within the SVS Core and each SVS Sync's `numFetches`.
- Utilize and build off of a different implementation for `orderedmap`s. Reduces `StateVector` memory usage by half and improves performance for most operations including parsing.

### Added
- `orderedmap`s own implemenation of a list (not available from a API standpoint).

### Removed
- The generic list dependency due to its non-use.

## [v0.0.0-alpha.5] - 2022-11-19
### Changed
- StateVector encoding optimization, entry lengths are reused.
- Interfaced Scheduler, Core, and all Syncs within SVS.
- Consolidated small files in SVS.
- Other small changes.

### Security
- Eliminated 6+ (all that are known) data races found in SVS.

## [v0.0.0-alpha.4] - 2022-11-13
### Added
- All Syncs in SVS now implement retries!
- BloomFilter code. (for future plans)

### Changed
- Standardize the seqno within SVS to a uint64.
- Utilize go-ndn's methods for encoding.
- Exposed all internal through util due to necessary access.
- Modified to ensure compatibility to go-ndn's latest changes.

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

[Unreleased]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.13...HEAD
[v0.0.0-alpha.13]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.12...v0.0.0-alpha.13
[v0.0.0-alpha.12]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.11...v0.0.0-alpha.12
[v0.0.0-alpha.11]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.10...v0.0.0-alpha.11
[v0.0.0-alpha.10]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.9...v0.0.0-alpha.10
[v0.0.0-alpha.9]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.8...v0.0.0-alpha.9
[v0.0.0-alpha.8]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.7...v0.0.0-alpha.8
[v0.0.0-alpha.7]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.6...v0.0.0-alpha.7
[v0.0.0-alpha.6]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.5...v0.0.0-alpha.6
[v0.0.0-alpha.5]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.4...v0.0.0-alpha.5
[v0.0.0-alpha.4]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.3...v0.0.0-alpha.4
[v0.0.0-alpha.3]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.2...v0.0.0-alpha.3
[v0.0.0-alpha.2]: https://github.com/justincpresley/ndn-sync/compare/v0.0.0-alpha.1...v0.0.0-alpha.2
[v0.0.0-alpha.1]: https://github.com/justincpresley/ndn-sync/releases/tag/v0.0.0-alpha.1
