module github.com/Necoro/feed2imap-go

go 1.16

require (
	github.com/PuerkitoBio/goquery v1.7.0
	github.com/antonmedv/expr v1.8.9
	github.com/emersion/go-imap v1.1.0
	github.com/emersion/go-imap-uidplus v0.0.0-20200503180755-e75854c361e9
	github.com/emersion/go-message v0.15.0
	github.com/gabriel-vasile/mimetype v1.3.1
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/mmcdole/gofeed v1.1.3
	github.com/nightlyone/lockfile v1.0.0
	golang.org/x/net v0.0.0-20210505024714-0287a6fb4125
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/jaytaylor/html2text => github.com/Necoro/html2text v0.0.0-20200822202223-2e8e4cfbb241
	github.com/mmcdole/gofeed => github.com/Necoro/gofeed v1.1.1-0.20210423205404-5a9d204f8125
)
