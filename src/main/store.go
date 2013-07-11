package main

import (
	"github.com/peterbourgon/diskv"
)

// holds Diskv database
var store *diskv.Diskv

func initStore(cfg Config) {
	// Initialize a new diskv store, rooted at "store", with a 1MB cache.
	store = diskv.New(diskv.Options{
		BasePath:     cfg.Server.Store,
		Transform:    BlockTransform,
		CacheSizeMax: 1024 * 1024,
	})
}
