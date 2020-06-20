package feed

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Version byte

const (
	currentVersion Version = 1
)

type Cache interface {
	findItem(*Feed) CachedFeed
	Version() Version
	Info() string
	SpecificInfo(interface{}) string
	transformToCurrent() (Cache, error)
}

type CachedFeed interface {
	Checked(withFailure bool)
	Failures() int
	Last() time.Time
	ID() string
	filterItems(items []item, ignoreHash bool, alwaysNew bool) []item
	Commit()
}

func cacheForVersion(version Version) (Cache, error) {
	switch version {
	case v1Version:
		return newV1Cache(), nil
	default:
		return nil, fmt.Errorf("unknown cache version '%d'", version)
	}
}

func storeCache(cache Cache, fileName string) error {
	if cache == nil {
		return fmt.Errorf("trying to store nil cache")
	}
	if cache.Version() != currentVersion {
		return fmt.Errorf("trying to store cache with unsupported version '%d' (current: '%d')", cache.Version(), currentVersion)
	}

	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("trying to store cache to '%s': %w", fileName, err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	if err = writer.WriteByte(byte(currentVersion)); err != nil {
		return fmt.Errorf("writing to '%s': %w", fileName, err)
	}

	encoder := gob.NewEncoder(writer)
	if err = encoder.Encode(cache); err != nil {
		return fmt.Errorf("encoding cache: %w", err)
	}

	writer.Flush()
	log.Printf("Stored cache to '%s'.", fileName)

	return nil
}

func newCache() (Cache, error) {
	return cacheForVersion(currentVersion)
}

func LoadCache(fileName string) (Cache, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no cache there yet -- make new
			return newCache()
		}
		return nil, fmt.Errorf("opening cache at '%s': %w", fileName, err)
	}
	defer f.Close()

	log.Printf("Loading cache from '%s'", fileName)

	reader := bufio.NewReader(f)
	version, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading from '%s': %w", fileName, err)
	}

	cache, err := cacheForVersion(Version(version))
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(reader)
	if err = decoder.Decode(cache); err != nil {
		return nil, fmt.Errorf("decoding for version '%d' from '%s': %w", version, fileName, err)
	}

	if cache, err = cache.transformToCurrent(); err != nil {
		return nil, fmt.Errorf("cannot transform from version %d to %d: %w", version, currentVersion, err)
	}

	log.Printf("Loaded cache (version %d), transformed to version %d.", version, currentVersion)

	return cache, nil
}
