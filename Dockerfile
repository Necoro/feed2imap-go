FROM alpine

RUN mkdir /app
WORKDIR /app
COPY dist/feed2imap-go_linux_amd64/feed2imap-go /app/feed2imap-go
COPY config.yml /app/config.yml

ENTRYPOINT ["/app/feed2imap-go", "-c", "/app/data/feed.cache", "-f", "/app/config.yml" ]

