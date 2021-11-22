package cache

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/nightlyone/lockfile"

	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Version byte

const (
	currentVersion Version = v2Version
)

type Impl interface {
	cachedFeed(*feed.Feed) CachedFeed
	transformTo(Version) (Impl, error)
	cleanup(knownDescriptors map[feed.Descriptor]struct{})
	load(io.Reader) error
	store(io.Writer) error
	Version() Version
	Info() string
	SpecificInfo(interface{}) string
}

type Cache struct {
	Impl
	lock   lockfile.Lockfile
	locked bool
}

type CachedFeed interface {
	// Checked marks the feed as being a failure or a success on last check.
	Checked(withFailure bool)
	// Failures of this feed up to now.
	Failures() int
	// The Last time, this feed has been checked
	Last() time.Time
	// Filter the given items against the cached items.
	Filter(items []feed.Item, ignoreHash bool, alwaysNew bool) []feed.Item
	// Commit any changes done to the cache state.
	Commit()
	// The Feed, that is cached.
	Feed() *feed.Feed
}

func forVersion(version Version) (Impl, error) {
	switch version {
	case v1Version:
		return newV1Cache(), nil
	case v2Version:
		return newV2Cache(), nil
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
	if cache.Impl == nil {
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

	if err = cache.Impl.store(writer); err != nil {
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

func create() (Cache, error) {
	cache, err := forVersion(currentVersion)
	if err != nil {
		return Cache{}, err
	}
	return Cache{
		Impl:   cache,
		locked: false,
	}, nil
}

func Load(fileName string, upgrade bool) (Cache, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no cache there yet -- make new
			return create()
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

	cache, err := forVersion(Version(version))
	if err != nil {
		return Cache{}, err
	}

	if err = cache.load(reader); err != nil {
		return Cache{}, fmt.Errorf("decoding for version '%d' from '%s': %w", version, fileName, err)
	}

	if upgrade && currentVersion != cache.Version() {
		if cache, err = cache.transformTo(currentVersion); err != nil {
			return Cache{}, fmt.Errorf("cannot transform from version %d to %d: %w", version, currentVersion, err)
		}

		log.Printf("Loaded cache (version %d), transformed to version %d.", version, currentVersion)
	} else {
		log.Printf("Loaded cache (version %d)", version)
	}

	return Cache{cache, lock, true}, nil
}
