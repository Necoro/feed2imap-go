package cache

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/Necoro/feed2imap-go/internal/feed"
)

const v2Version Version = 2

// v2Cache is identical to v1Cache, but uses gzip compression for storage
type v2Cache v1Cache

func newV2Cache() *v2Cache {
	return (*v2Cache)(newV1Cache())
}

func (cache *v2Cache) asV1() *v1Cache {
	return (*v1Cache)(cache)
}

func (cache *v2Cache) cachedFeed(feed *feed.Feed) CachedFeed {
	return cache.asV1().cachedFeed(feed)
}

func (cache *v2Cache) transformTo(v Version) (Impl, error) {
	return nil, fmt.Errorf("Transformation not supported")
}

func (cache *v2Cache) cleanup(knownDescriptors map[feed.Descriptor]struct{}) {
	cache.asV1().cleanup(knownDescriptors)
}

func (cache *v2Cache) Version() Version {
	return v2Version
}

func (cache *v2Cache) Info() string {
	return cache.asV1().Info()
}

func (cache *v2Cache) SpecificInfo(i interface{}) string {
	return cache.asV1().SpecificInfo(i)
}

func (cache *v2Cache) load(reader io.Reader) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	return cache.asV1().load(gzipReader)
}

func (cache *v2Cache) store(writer io.Writer) error {
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	if err := cache.asV1().store(gzipWriter); err != nil {
		return err
	}

	return gzipWriter.Flush()
}
