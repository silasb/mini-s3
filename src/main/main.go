package main

import (
	"fmt"
	"github.com/studygolang/mux"
	"net/http"
	"code.google.com/p/gcfg"
	"filter"
)

func main() {
	fmt.Println("mini-s3 v0.0.1")

	var cfg Config

	// setting default settings
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = "8080"
	cfg.Server.DomainName = "s3.dev"
	cfg.Server.Store = "store"

	err := gcfg.ReadFileInto(&cfg, "config")
	if err != nil {
		fmt.Println("Not using config file for settings, using defaults.")
	}

	// these two functions initialize global variables in their respective files
	initStore(cfg)
	initMetaStore()

	bucketFilter := new(filter.BucketFilter)
	bucketFilterChain := mux.NewFilterChain(bucketFilter)

	r := mux.NewRouter()

	s := r.Host(`{subdomain}.` + cfg.Server.DomainName).Subrouter()
	// check if the bucket is included or not.
	s.FilterChain(bucketFilterChain)

	s.HandleFunc("/", BucketHandler)
	s.HandleFunc(`/{object:[a-zA-Z0-9_/\.]+}`, GETObjectHandler).Methods("GET")
	s.HandleFunc(`/{object:[a-zA-Z0-9_/\.]+}`, POSTObjectHandler).Methods("POST")
	s.HandleFunc(`/{object:[a-zA-Z0-9_/\.]+}`, DeleteObjectHandler).Methods("DELETE")
	http.Handle("/", s)

	http.ListenAndServe(cfg.Server.Host+":"+cfg.Server.Port, s)
}
