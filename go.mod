module github.com/Necoro/feed2imap-go

go 1.15

require (
	github.com/Necoro/html2text v0.0.0-20200510210108-d7611f0be99f
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/antonmedv/expr v1.8.8
	github.com/emersion/go-imap v1.0.5
	github.com/emersion/go-imap-uidplus v0.0.0-20200503180755-e75854c361e9
	github.com/emersion/go-message v0.11.3-0.20200422153910-8c6ac6b57e3d
	github.com/gabriel-vasile/mimetype v1.1.1
	github.com/google/go-cmp v0.5.1
	github.com/google/uuid v1.1.1
	github.com/mmcdole/gofeed v1.0.0
	github.com/nightlyone/lockfile v1.0.0
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f
	gopkg.in/yaml.v3 v3.0.0-20200506231410-2ff61e1afc86
)

replace github.com/mmcdole/gofeed => github.com/Necoro/gofeed v1.0.1-0.20200822192128-d9090592eb44
