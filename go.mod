module github.com/Necoro/feed2imap-go

go 1.17

require (
	github.com/PuerkitoBio/goquery v1.7.1
	github.com/antonmedv/expr v1.9.0
	github.com/emersion/go-imap v1.2.0
	github.com/emersion/go-imap-uidplus v0.0.0-20200503180755-e75854c361e9
	github.com/emersion/go-message v0.15.0
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/mmcdole/gofeed v1.1.3
	github.com/nightlyone/lockfile v1.0.0
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/andybalholm/cascadia v1.2.0 // indirect
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21 // indirect
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace (
	github.com/jaytaylor/html2text => github.com/Necoro/html2text v0.0.0-20210724110643-65369e0955db
	github.com/mmcdole/gofeed => github.com/Necoro/gofeed v1.1.1-0.20210423205404-5a9d204f8125
)
