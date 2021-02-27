package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Necoro/feed2imap-go/internal/feed/cache"
)

// flags
var (
	cacheFile string = "feed.cache"
	feedId    string = ""
)

func init() {
	flag.StringVar(&cacheFile, "c", cacheFile, "cache file")
	flag.StringVar(&feedId, "i", feedId, "id of the feed")
}

func main() {
	flag.Parse()

	cache, err := cache.Load(cacheFile, false)
	if err != nil {
		log.Fatal(err)
	}

	defer cache.Unlock()

	fmt.Printf("Cache version %d\n", cache.Version())
	if feedId != "" {
		fmt.Print(cache.SpecificInfo(feedId))
	} else {
		fmt.Print(cache.Info())
	}
}
