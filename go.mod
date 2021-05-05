module github.com/Necoro/feed2imap-go

go 1.16

require (
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/antonmedv/expr v1.8.9
	github.com/emersion/go-imap v1.0.7-0.20201107145532-de4b254f4d93
	github.com/emersion/go-imap-uidplus v0.0.0-20200503180755-e75854c361e9
	github.com/emersion/go-message v0.14.1
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/mmcdole/gofeed v1.1.3
	github.com/nightlyone/lockfile v1.0.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace (
	github.com/jaytaylor/html2text => github.com/Necoro/html2text v0.0.0-20200822202223-2e8e4cfbb241
	github.com/mmcdole/gofeed => github.com/Necoro/gofeed v1.1.1-0.20210423205404-5a9d204f8125
)
