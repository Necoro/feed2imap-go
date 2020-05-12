FROM alpine

RUN mkdir /app
COPY feed2imap-go /app

ENTRYPOINT ["/app/feed2imap-go"]
CMD ["-c", "/app/data/feed.cache", "-f", "/app/data/config.yml"]
