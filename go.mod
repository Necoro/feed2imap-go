module github.com/Necoro/feed2imap-go

go 1.14

require (
	github.com/emersion/go-imap v1.0.4
	github.com/emersion/go-message v0.11.2
	github.com/mmcdole/gofeed v1.0.0-beta2.0.20200331235650-4298e4366be3
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

replace github.com/emersion/go-message => github.com/Necoro/go-message v0.11.3-pre
