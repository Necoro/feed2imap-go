FROM alpine

COPY feed2imap-go /

ENTRYPOINT ["/feed2imap-go"]
CMD ["-c", "/data/feed.cache", "-f", "/data/config.yml"]
