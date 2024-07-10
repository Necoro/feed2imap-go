# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.7.2] - 2024-07-10
### Fixed
- Cache loading problem introduced by v1.7.1

## [1.7.1] - 2024-07-10
- Upgraded dependencies

## [1.7.0] - 2023-06-13
### Fixed
- [Issue #97](https://github.com/Necoro/feed2imap-go/issues/97): Fix panic when an IMAP connection is tried to established when the program has already entered the `Disconnect` state.
### Added
- New global config variable `max-imap-connections`, with a default of `5`.
### Changed
- Change go-readability back to track upstream. This now incorporates an overhaul of the readability engine, so changes are to be expected when using `body: fetch`.
- Improve strictness regarding config parsing, especially in the differentiation between feeds and groups of feeds.
- [Issue #95](https://github.com/Necoro/feed2imap-go/issues/95): Ensure cache is not world-readable. Add a warning when the config is world-readable.

## [1.6.0] - 2023-03-28
- Upgrade dependencies
### Fixed
- [Issue #91](https://github.com/Necoro/feed2imap-go/issues/91): Panic on using per-item targets

## [1.5.2] - 2023-01-31
- Upgrade dependencies

## [1.5.1] - 2022-06-11
- Upgrade dependencies
- Minor lifting to Go 1.18

## [1.5.0] - 2022-01-11
### Added
- [Issue #66](https://github.com/Necoro/feed2imap-go/issues/66): Allow specifying custom template files for HTML/Text output (configuration: `html-template`/`text-template`).
- [Issue #67](https://github.com/Necoro/feed2imap-go/issues/67): Support for fetching the linked article/website instead of using the body in the feed (configuration: `body: fetch`). This is especially useful when the feed only supplies links/teaser and not the full content.
### Fixed
- Panic on setting `body` on feed-level.

## [1.4.0] - 2021-12-21
### Added
- Make links absolute: relative links inside a feed cannot be resolved outside a webbrowser (Example/culprit: https://go.dev/blog/go1.18beta1)

## [1.3.0] - 2021-11-01
### Added
- [Issue #64](https://github.com/Necoro/feed2imap-go/issues/64): Set filename for included images.

## [1.2.0] - 2021-10-20
### Added
- Location of the cache can now be specified in the config.yml. Rationale: Easier way of dealing with multiple configurations, as each also requires a distinct cache.
- [Issue #6](https://github.com/Necoro/feed2imap-go/issues/6): Support old-style configurations with imap targets on each feed. Restriction: Servers must be equal over all connection strings!
### Fixed
- [Issue #62](https://github.com/Necoro/feed2imap-go/issues/62): Allow empty root folder.

## [1.1.1] - 2021-07-24
### Fixed
- Correctly log out from the imap server and do not harshly disconnect.

## [1.1.0] - 2021-06-18
### Fixed
- Do not try to download already embedded images (i.e. `data:image/`).
### Changed
- Updated dependencies. Most notable: Upgrade of yaml.v3, which entails changes of nil-handling.

## [1.0.0] - 2021-05-19
### Fixed
- [Issue #47](https://github.com/Necoro/feed2imap-go/issues/47): Fixed occassional deadlocks. Reason was unilateral updates from the server which were ill-handled by go-imap.

## [0.8.0] - 2021-03-06
### Added
- New cache format v2 that uses gzip compression
- Support for JSON v1.1 feed (via [gofeed](https://github.com/mmcdole/gofeed/pull/169))
### Changed
- Caches store now 1000 old entries (i.e., not included in the last fetch) at maximum. This will clean obsolete cruft and drastically reduce cache size.
- Feeds not updated (and not part of the config) for 180 days are now automatically removed from the cache.
- Connecting to the IMAP server now happens in the background and in parallel, using connections directly as they are established. This should yield a speed-up with slow IMAP servers.
- IMAP connections now can time out during establishment.

## [0.7.0] - 2021-02-21
### Changed
- Remove `srcset` attribute of `img` tags when including images in mail
- Strip whitespaces from folder names

### Fixed
- [Issue #39](https://github.com/Necoro/feed2imap-go/issues/39): Do not re-introduce deleted mails, even though `reupload-if-updated` is false.
- [Issue #25](https://github.com/Necoro/feed2imap-go/issues/25): Normalize folder names, so `foo` and `foo/` are not seen as different folders.

## [0.6.0] - 2021-02-14
### Fixed
- [Issue #46](https://github.com/Necoro/feed2imap-go/issues/46): Fixed line endings in templates, thereby pleasing Cyrus IMAP server.

## [0.5.2] - 2020-11-23
- Update of libraries
- This now also includes the updated gofeed dependency, that was promised with the last version but not included...

## [0.5.1] - 2020-09-11
- Update of gofeed dependency: Now supports json feeds
- Make sure, cache locks are deleted on shutdown (looks cleaner)

## [0.5.0] - 2020-08-22
### Added
- Cache files are now explicitly locked. This avoids two instances of feed2imap-go running at the same time.
- New header `X-Feed2Imap-Create-Date` holding the date of creation of that mail. Mostly needed for debugging issues.
- Updated to Go 1.15.
- New global option `auto-target` to change the default behavior of omitted `target` fields.  
  When set to `false`, a missing `target` is identical to specifying `null` or `""` for the target.  
  When set to `true` (the default), the standard behavior of falling back onto the name is used.

### Fixed
- [Issue #24](https://github.com/Necoro/feed2imap-go/issues/24): Patched gofeed to support atom tags in RSS feeds

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

[Unreleased]: https://github.com/Necoro/feed2imap-go/compare/v1.7.2...HEAD
[1.7.2]: https://github.com/Necoro/feed2imap-go/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/Necoro/feed2imap-go/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/Necoro/feed2imap-go/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/Necoro/feed2imap-go/compare/v1.5.2...v1.6.0
[1.5.2]: https://github.com/Necoro/feed2imap-go/compare/v1.5.1...v1.5.2
[1.5.1]: https://github.com/Necoro/feed2imap-go/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/Necoro/feed2imap-go/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/Necoro/feed2imap-go/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/Necoro/feed2imap-go/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/Necoro/feed2imap-go/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/Necoro/feed2imap-go/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/Necoro/feed2imap-go/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/Necoro/feed2imap-go/compare/v0.8.0...v1.0.0
[0.8.0]: https://github.com/Necoro/feed2imap-go/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/Necoro/feed2imap-go/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/Necoro/feed2imap-go/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/Necoro/feed2imap-go/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/Necoro/feed2imap-go/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/Necoro/feed2imap-go/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/Necoro/feed2imap-go/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/Necoro/feed2imap-go/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/Necoro/feed2imap-go/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/Necoro/feed2imap-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Necoro/feed2imap-go/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Necoro/feed2imap-go/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Necoro/feed2imap-go/releases/tag/v0.1.0
