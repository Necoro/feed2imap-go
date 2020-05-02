[![Go Report Card](https://goreportcard.com/badge/github.com/Necoro/feed2imap-go)](https://goreportcard.com/report/github.com/Necoro/feed2imap-go)

# feed2imap-go

A software to convert rss feeds into mails. feed2imap-go acts an an RSS/Atom feed aggregator. After downloading feeds 
(over HTTP or HTTPS), it uploads them to a specified folder of an IMAP mail server. The user can then access the feeds 
using their preferred client (Mutt, Evolution, Mozilla Thunderbird, webmail,...).

It is a rewrite in Go of the wonderful, but unfortunately now unmaintained, [feed2imap](https://github.com/feed2imap/feed2imap).
It also includes the features that up to now only lived on [my own branch][nec].

It aims to be compatible in functionality and configuration, and should mostly work as a drop-in replacement 
(but see [Changes](#changes)).

## Features

* Support for most feed formats. See [gofeed documentation](https://github.com/mmcdole/gofeed/blob/master/README.md#features) 
for details.
* Connection to any IMAP server, using IMAP, IMAP+STARTTLS, or IMAPS.
* Detection of duplicates: Heuristics what feed items have already been uploaded.
* Update mechanism: When a feed item is updated, so is the mail. (_TODO_: [issue #9][i9])
* Detailed configuration options per feed (fetch frequency, should images be included, tune change heuristics, ...)

## Changes

### Additions to feed2imap

* groups (_details TBD_)
* heavier use of parallel processing (it's Go after all ;))
* Global `target` and each feed only specifies the folder relative to that target. 
(feature contained also in [fork of the original][nec]) 
* Fix `include-images` option: It now includes images as mime-parts. An additional `embed-images` option serves the images 
as inline base64-encoded data (the old default behavior of feed2imap).
* Improved image inclusion: Links without scheme; images without extension (using mime-detection)
* Use HTML-Parser instead of regular expressions for modifying the HTML content.

### Subtle differences

* **Feed rendering**: Unfortunately, semantics of RSS and Atom tags are very broad. As we use a different feed parser 
ibrary than the original, the interpretation (e.g., what tag is "the author") can differ.
* **Caching**: We do not implement the caching algorithm of feed2imap point by point. In general we opted for less 
heuristics and more optimism (belief that GUID is filled correctly; belief that the difference between publishing and 
update date is adhered to). If this results in a problem, file a bug and include the `X-Feed2Imap-Reason` header of the mail.
* **Configuration**: We took the liberty to restructure the configuration options. Old configs are supported, but a 
warning is issued when an option should now be in another place or is no longer supported (i.e., the option is without function).

### Unsupported features of feed2imap

* IMAP-Target per Feed ([issue #6][i6]); targets only specify the folder relative to the global target
* Maildir ([issue #4][i4])
* Scripts for generating/filtering feeds

[i6]: https://github.com/Necoro/feed2imap-go/issues/6
[i4]: https://github.com/Necoro/feed2imap-go/issues/4
[i9]: https://github.com/Necoro/feed2imap-go/issues/9
[nec]: https://github.com/Necoro/feed2imap
