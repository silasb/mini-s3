package main

import (
	"github.com/jmhodges/levigo"
	"fmt"
)

var meta_store *levigo.DB

func initMetaStore() {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(3<<30))
	opts.SetCreateIfMissing(true)

	var err error

	meta_store, err = levigo.Open("./meta_store", opts)

	if err != nil {
		fmt.Println(err)
		return
	}
}