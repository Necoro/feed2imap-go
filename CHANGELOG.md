# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- New header X-Feed2Imap-Guid

### Changed
- Default for `min-frequency` is now 0 instead of 1.
- Fixed date parsing from feed. _Changes cache format!_
- Misc bug fixes in template rendering.

## [0.1.1] - 2020-05-04

### Added
- Automatic releasing via goreleaser

### Changed
- Improved version output

## [0.1.0] - 2020-05-04

Initial release

[Unreleased]: https://github.com/Necoro/feed2imap-go/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/Necoro/feed2imap-go/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Necoro/feed2imap-go/releases/tag/v0.1.0
