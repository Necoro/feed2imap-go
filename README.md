[![Go Report Card](https://goreportcard.com/badge/github.com/Necoro/feed2imap-go)](https://goreportcard.com/report/github.com/Necoro/feed2imap-go)

# feed2imap-go

A software to convert rss feeds into mails. feed2imap-go acts as an RSS/Atom feed aggregator. After downloading feeds 
(over HTTP or HTTPS), it uploads them to a specified folder of an IMAP mail server. The user can then access the feeds 
using their preferred client (Mutt, Evolution, Mozilla Thunderbird, webmail,...).

It is a rewrite in Go of the wonderful, but unfortunately now unmaintained, [feed2imap](https://github.com/feed2imap/feed2imap).
It also includes the features that up to now only lived on [my own branch][nec].

It aims to be compatible in functionality and configuration, and should mostly work as a drop-in replacement 
(but see [Changes](#changes)).

An example configuration can be found [here](config.yml.example).

See [the Installation section](#installation) on how to install feed2imap-go. (Spoiler: It's easy ;)).

## Features

* Support for most feed formats. See [gofeed documentation](https://github.com/mmcdole/gofeed/blob/master/README.md#features) 
for details.
* Connection to any IMAP server, using IMAP, IMAP+STARTTLS, or IMAPS.
* Detection of duplicates: Heuristics what feed items have already been uploaded.
* Update mechanism: When a feed item is updated, so is the mail.
* Detailed configuration options per feed (fetch frequency, should images be included, tune change heuristics, ...)

## Changes

### Additions to feed2imap

* groups (_details TBD_)
* Heavier use of parallel processing (it's Go after all ;)). Also, it is way faster.
* Global `target` and each feed only specifies the folder relative to that target. 
(feature contained also in [fork of the original][nec]) 
* Fix `include-images` option: It now includes images as mime-parts. An additional `embed-images` option serves the images 
as inline base64-encoded data (the old default behavior of feed2imap).
* Improved image inclusion: Support any relative URLs, including `//example.com/foo.png`
* Use HTML-Parser instead of regular expressions for modifying the HTML content.
* STARTTLS-Support. As it turned out only in testing, the old feed2imap never supported it...

### Subtle differences

* **Feed rendering**: Unfortunately, semantics of RSS and Atom tags are very broad. As we use a different feed parser 
ibrary than the original, the interpretation (e.g., what tag is "the author") can differ.
* **Caching**: We do not implement the caching algorithm of feed2imap point by point. In general, we opted for fewer 
heuristics and more optimism (belief that GUID is filled correctly). If this results in a problem, file a bug and include the `X-Feed2Imap-Reason` header of the mail.
* **Configuration**: We took the liberty to restructure the configuration options. Old configs are supported, but a 
warning is issued when an option should now be in another place or is no longer supported (i.e., the option is without function).

### Unsupported features of feed2imap

* IMAP-Target per Feed ([issue #6][i6]); targets only specify the folder relative to the global target
* Maildir ([issue #4][i4])
* Scripts for generating/filtering feeds

## Installation

The easiest way of installation is to head over to [the releases page](https://github.com/Necoro/feed2imap-go/releases/latest)
and get the appropriate download package. Go is all about static linking, thus for all platforms the result is a single
binary which can be placed whereever you need.

Please open an issue if you are missing your platform.

### Install from source

Clone the repository and, optionally, switch to the tag you want:
````bash
git clone https://github.com/Necoro/feed2imap-go
git checkout v0.1.1
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
go get https://github.com/Necoro/feed2imap-go
````
Note that this is not recommended, because it is a) not obvious which version you are actually building
and b) the update process is not clear, especially regarding the dependencies.


### Run in docker

Build docker image:

````bash
docker build -t feed2imap-go .
````

Run docker-command:

````bash
docker volume create docker_feeddata
docker container run -v feeddata:/app/data feed2imap-go
````

Or run with docker-compose:

````bash
docker-compose up

To use the Docker image on other servers push the image into a docker registry

````bash
docker build -t docker.your.domain/feed2imap .

docker login <your.docker.registry.url>

docker push docker.your.domain/feed2imap
````

Use in docker-compose `FROM docker.your.domain/feed2imap` in kubernetes yaml `image: docker.your.domain/feed2imap` 

Deploy to Kubernetes as Cronjob. 

Change kubernetes/cronjob.yaml with your data. 

````bash
kubectl apply -f kubernetes/cronjob.yaml
configmap/config-yml created
cronjob.batch/feed2imap created
persistentvolume/feed2imap-pv-volume created
persistentvolumeclaim/feed2imap-pv-claim created
````



Next Todo: Run scheduled kubernetes job in k8s cluster 

[i6]: https://github.com/Necoro/feed2imap-go/issues/6
[i4]: https://github.com/Necoro/feed2imap-go/issues/4
[i9]: https://github.com/Necoro/feed2imap-go/issues/9
[nec]: https://github.com/Necoro/feed2imap
