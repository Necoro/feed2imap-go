package cache

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	"github.com/Necoro/feed2imap-go/internal/log"
)

const currentVersion byte = 1

type Cache interface {
	Version() byte
	transformToCurrent() (Cache, error)
}

type feedId struct {
	Name string
	Url  string
}

type v1Cache struct {
	version byte
	Ids     map[feedId]uint64
	NextId  uint64
}

func (cache *v1Cache) Version() byte {
	return cache.version
}

func New() Cache {
	cache := v1Cache{Ids: map[feedId]uint64{}}
	cache.version = currentVersion
	return &cache
}

func cacheForVersion(version byte) (Cache, error) {
	switch version {
	case 1:
		return New(), nil
	default:
		return nil, fmt.Errorf("unknown cache version '%d'", version)
	}
}

func (cache *v1Cache) transformToCurrent() (Cache, error) {
	return cache, nil
}

func Store(fileName string, cache Cache) error {
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
	if err = writer.WriteByte(currentVersion); err != nil {
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

func Read(fileName string) (Cache, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no cache there yet -- make new
			return New(), nil
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

	cache, err := cacheForVersion(version)
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
