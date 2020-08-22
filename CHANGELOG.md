# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Cache files are now explicitly locked. This avoids two instances of feed2imap-go running at the same time.
- Updated to Go 1.15.
- New global option `auto-target` to change the default behavior of omitted `target` fields.  
  When set to `false`, a missing `target` is identical to specifying `null` or `""` for the target.  
  When set to `true` (the default), the standard behavior of falling back onto the name is used.

## [0.4.1] - 2020-06-20
### Fixed
- Fix a bug, where cached items get deleted when a feed runs dry. 
This resulted in duplicate entries as soon as the feed contained (possibly older) entries again. 

## [0.4.0] - 2020-05-25
### Added
- Verbose variant of 'target' in config: Do not hassle with urlencoded passwords anymore!
- New feed option 'item-filter' for filtering out specific items from feed.
- New feed option 'exec', allowing to specify a command to execute instead of a Url to fetch from.

## [0.3.1] - 2020-05-12
- Docker Setup

## [0.3.0] - 2020-05-10
### Added
- Options are now also allowed on group-level (closes #12)
- Render text parts in mails (closes #7)

## [0.2.0] - 2020-05-10
### Added
- New header X-Feed2Imap-Guid

### Changed
- Default for `min-frequency` is now 0 instead of 1.
- Fixed date parsing from feed. _Changes cache format!_
- Do not assume items to be new when their published date is newer than the last run. Some feeds just lie...
- Misc bug fixes in template rendering.

## [0.1.1] - 2020-05-04

### Added
- Automatic releasing via goreleaser

### Changed
- Improved version output

## [0.1.0] - 2020-05-04

Initial release

[Unreleased]: https://github.com/Necoro/feed2imap-go/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/Necoro/feed2imap-go/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/Necoro/feed2imap-go/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/Necoro/feed2imap-go/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/Necoro/feed2imap-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Necoro/feed2imap-go/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Necoro/feed2imap-go/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Necoro/feed2imap-go/releases/tag/v0.1.0
