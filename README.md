[![Go Report Card](https://goreportcard.com/badge/github.com/Necoro/feed2imap-go)](https://goreportcard.com/report/github.com/Necoro/feed2imap-go)

# feed2imap-go

A software to convert rss feeds into mails. feed2imap-go acts as an RSS/Atom feed aggregator. After downloading feeds 
(over HTTP or HTTPS), it uploads them to a specified folder of an IMAP mail server. The user can then access the feeds 
using their preferred client (Mutt, Evolution, Mozilla Thunderbird, webmail,...).

It is a rewrite in Go of the wonderful, but unfortunately now unmaintained, [feed2imap](https://github.com/feed2imap/feed2imap).
It also includes the features that up to now only lived on [my own branch][nec].

It aims to be compatible in functionality and configuration, and should mostly work as a drop-in replacement 
(but see [Changes](#changes)).

An example configuration can be found [here](config.yml.example) with additional information in the [wiki](https://github.com/Necoro/feed2imap-go/wiki/Detailed-Options).

See [the Installation section](#installation) on how to install feed2imap-go. (Spoiler: It's easy ;)).

## Features

* Support for most feed formats. See [gofeed documentation](https://github.com/mmcdole/gofeed/blob/master/README.md#features) 
for details. Feeds need not be supplied via URL but can also be yielded by an executable.
* Connection to any IMAP server, using IMAP, IMAP+STARTTLS, or IMAPS.
* Detection of duplicates: Heuristics what feed items have already been uploaded.
* Update mechanism: When a feed item is updated, so is the mail.
* Detailed configuration options per feed (fetch frequency, should images be included, tune change heuristics, ...)
* Support for custom filters on feeds
* [Readability support][i67], i.e. fetching and presenting the linked content instead of the teaser/link included in the feed itself.

## Changes

### Additions to feed2imap

* Groups: Have the ability to group feeds that share characteristics, most often the same parent folder in the hiearchy.
  It also allows sharing options between feeds. Usages: Categories ("News", "Linux") and merging different feeds of the same origin.
* Heavier use of parallel processing (it's Go after all ;)). Also, it is way faster.
* Global `target` and each feed only specifies the folder relative to that target. 
(feature contained also in [fork of the original][nec]) 
* Fix `include-images` option: It now includes images as mime-parts. An additional `embed-images` option serves the images 
as inline base64-encoded data (the old default behavior of feed2imap).
* Improved image inclusion: Support any relative URLs, including `//example.com/foo.png`
* Use an HTML parser instead of regular expressions for modifying the HTML content.
* STARTTLS support. As it turned out only in testing, the old feed2imap never supported it...
* `item-filter` option that allows to specify an inline filter expression on the items of a feed.
* Readability support: Fetch and present the linked article.
* Mail templates can be customized.

### Subtle differences

* **Feed rendering**: Unfortunately, semantics of RSS and Atom tags are very broad. As we use a different feed parser 
ibrary than the original, the interpretation (e.g., what tag is "the author") can differ.
* **Caching**: We do not implement the caching algorithm of feed2imap point by point. In general, we opted for fewer 
heuristics and more optimism (belief that GUID is filled correctly). If this results in a problem, file a bug and include the `X-Feed2Imap-Reason` header of the mail.
* **Configuration**: We took the liberty to restructure the configuration options. Old configs are supported, but a 
warning is issued when an option should now be in another place or is no longer supported (i.e., the option is without function).

### Unsupported features of feed2imap

* Maildir ([issue #4][i4])
* Different IMAP servers in the same configuration file. Please use multiple config files, if this is needed (see also [issue #6][i6]).

## Installation

The easiest way of installation is to head over to [the releases page](https://github.com/Necoro/feed2imap-go/releases/latest)
and get the appropriate download package. Go is all about static linking, thus for all platforms the result is a single
binary which can be placed whereever you need.

Please open an issue if you are missing your platform.

### Install from source

Clone the repository and, optionally, switch to the tag you want:
````bash
git clone https://github.com/Necoro/feed2imap-go
git checkout v1.5.0
````

The official way of building feed2imap-go is using [goreleaser](https://github.com/goreleaser/goreleaser):
````bash
goreleaser --snapshot --rm-dist
````
The built binary is then inside the corresponding arch folder in `dist`.

In case you do not want to install yet another build tool, doing
````bash
go build
````
should also suffice, but does not embed version information in the binary (and the result is slightly larger).

If you are only interested in getting the latest build out of the `master` branch, do
````bash
go install github.com/Necoro/feed2imap-go@master
````
Using `@latest` instead of `@master` gives you the latest stable version.

### Run in docker

Most times, putting feed2imap-go somewhere and adding a cron job does everything you need. For the times when it isn't, we provide docker containers for your convenience at [Github Packages](https://github.com/Necoro/feed2imap-go/packages) and at [Docker Hub](https://hub.docker.com/r/necorodm/feed2imap-go).

The container is configured to expect both config file and cache under `/app/data/`, thus needs it mounted there.
When both are stored in `~/feed`, you can do:
````bash
docker run -v ~/feed:/app/data necorodm/feed2imap-go:latest
````

Alternatively, build the docker image yourself (requires the `feed2imap-go` binary at toplevel):
````bash
docker build -t feed2imap-go .
````
Note that the supplied binary must not be linked to glibc, i.e. has to be built with `CGO_ENABLED=0`. When using `goreleaser`, you'll find this in  `dist/docker_linux_amd64`.

Or you can roll your own Dockerfile, supplying a glibc...

**NB**: feed2imap-go employs no server-mode. Thus, each run terminates directly after a couple seconds. Therefore, the docker container in itself is not that useful, and you have to have a mechanism in place to spin up the container regularly.

## Support

Thanks to [JetBrains][jb] for supporting this project.

[![JetBrains](https://necoro.dev/data/jetbrains_small2.png)][jb]

[i6]: https://github.com/Necoro/feed2imap-go/issues/6
[i4]: https://github.com/Necoro/feed2imap-go/issues/4
[i9]: https://github.com/Necoro/feed2imap-go/issues/9
[i67]: https://github.com/Necoro/feed2imap-go/issues/67
[nec]: https://github.com/Necoro/feed2imap
[jb]: https://www.jetbrains.com/?from=feed2imap-go
