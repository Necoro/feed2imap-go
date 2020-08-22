package feed

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nightlyone/lockfile"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Version byte

const (
	currentVersion Version = 1
)

type CacheImpl interface {
	findItem(*Feed) CachedFeed
	Version() Version
	Info() string
	SpecificInfo(interface{}) string
	transformToCurrent() (CacheImpl, error)
}

type Cache struct {
	CacheImpl
	lock   lockfile.Lockfile
	locked bool
}

type CachedFeed interface {
	Checked(withFailure bool)
	Failures() int
	Last() time.Time
	ID() string
	filterItems(items []item, ignoreHash bool, alwaysNew bool) []item
	Commit()
}

func cacheForVersion(version Version) (CacheImpl, error) {
	switch version {
	case v1Version:
		return newV1Cache(), nil
	default:
		return nil, fmt.Errorf("unknown cache version '%d'", version)
	}
}

func lockName(fileName string) (string, error) {
	return filepath.Abs(fileName + ".lck")
}

func lock(fileName string) (lock lockfile.Lockfile, err error) {
	var lockFile string

	if lockFile, err = lockName(fileName); err != nil {
		return
	}
	log.Debugf("Handling lock file '%s'", lockFile)

	if lock, err = lockfile.New(lockFile); err != nil {
		err = fmt.Errorf("Creating lock file: %w", err)
		return
	}

	if err = lock.TryLock(); err != nil {
		err = fmt.Errorf("Locking cache: %w", err)
		return
	}

	return
}

func (cache *Cache) store(fileName string) error {
	if cache.CacheImpl == nil {
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
	if err = encoder.Encode(cache.CacheImpl); err != nil {
		return fmt.Errorf("encoding cache: %w", err)
	}

	writer.Flush()
	log.Printf("Stored cache to '%s'.", fileName)

	return cache.Unlock()
}

func (cache *Cache) Unlock() error {
	if cache.locked {
		if err := cache.lock.Unlock(); err != nil {
			return fmt.Errorf("Unlocking cache: %w", err)
		}
	}
	cache.locked = false
	return nil
}

func newCache() (Cache, error) {
	cache, err := cacheForVersion(currentVersion)
	if err != nil {
		return Cache{}, err
	}
	return Cache{
		CacheImpl: cache,
		locked:    false,
	}, nil
}

func LoadCache(fileName string) (Cache, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no cache there yet -- make new
			return newCache()
		}
		return Cache{}, fmt.Errorf("opening cache at '%s': %w", fileName, err)
	}
	defer f.Close()

	lock, err := lock(fileName)
	if err != nil {
		return Cache{}, err
	}

	log.Printf("Loading cache from '%s'", fileName)

	reader := bufio.NewReader(f)
	version, err := reader.ReadByte()
	if err != nil {
		return Cache{}, fmt.Errorf("reading from '%s': %w", fileName, err)
	}

	cache, err := cacheForVersion(Version(version))
	if err != nil {
		return Cache{}, err
	}

	decoder := gob.NewDecoder(reader)
	if err = decoder.Decode(cache); err != nil {
		return Cache{}, fmt.Errorf("decoding for version '%d' from '%s': %w", version, fileName, err)
	}

	if cache, err = cache.transformToCurrent(); err != nil {
		return Cache{}, fmt.Errorf("cannot transform from version %d to %d: %w", version, currentVersion, err)
	}

	log.Printf("Loaded cache (version %d), transformed to version %d.", version, currentVersion)

	return Cache{cache, lock, true}, nil
}
