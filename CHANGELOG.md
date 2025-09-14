# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2025-09-13

### Added

- `dump convert` command now converts to `ndjson` with the `--format ndjson` flag.

## [0.2.1] - 2025-09-12

### Fixes

- Removed unnecessary nested columns in parquet conversions.

## [0.2.0] - 2025-09-12

### Added

- `dump convert` command will convert a Discogs dump to a parquet file.
- `--stop-after=X` flag to `dump convert` to stop after X records.
- `--stop-after=X` flag to `dump structure` to stop after a X records.
- `CHANGELOG.md`.

### Changed

- Conversion or exports will set NULL values instead of default values (empty strings, 0, etc.)
