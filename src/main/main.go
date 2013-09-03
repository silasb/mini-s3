package main

import (
	"bitbucket.org/kardianos/osext"
	"code.google.com/p/gcfg"
	"filter"
	"fmt"
	"github.com/studygolang/mux"
	"net/http"
	"path"
	// uncomment for profile support
	//"github.com/davecheney/profile"
)

var current_exe_path string
var abs_store_path string

func main() {
	// uncomment for profile support
	//defer profile.Start(profile.CPUProfile).Stop()

	fmt.Println("mini-s3 v0.0.1")

	filename, _ := osext.Executable()
	current_exe_path = path.Dir(filename)

	var cfg Config

	// setting default settings
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = "8080"
	cfg.Server.RootDomainName = "s3.dev"
	cfg.Server.Store = "store"
	cfg.RPC.Host = "identity.api.pyserve.com"
	cfg.RPC.Port = 8001

	err := gcfg.ReadFileInto(&cfg, "config")
	if err != nil {
		fmt.Println("Not using config file for settings, using defaults.")
	}

	abs_store_path = path.Join(current_exe_path, cfg.Server.Store)

	// these two functions initialize global variables in their respective files
	initStore(cfg)
	initMetaStore()

	authFilter := new(filter.AuthFilter)
	authFilter.RPCServerHost = cfg.RPC.Host
	authFilter.RPCServerPort = cfg.RPC.Port
	bucketFilter := new(filter.BucketFilter)
	filterChange := mux.NewFilterChain([]mux.Filter{authFilter, bucketFilter}...)
	bucketFilterMux := mux.NewFilterChain(bucketFilter)

	r := mux.NewRouter()

	s := r.Host(`{subdomain}.` + cfg.Server.RootDomainName).Subrouter()

	s.HandleFunc("/", BucketHandler)
	s.HandleFunc(`/{object:[a-zA-Z0-9_/\.]+}`, GETObjectHandler).Methods("GET").AppendFilterChain(bucketFilterMux)
	s.HandleFunc(`/{object:[a-zA-Z0-9_/\.]+}`, POSTObjectHandler).Methods("POST").AppendFilterChain(filterChange)
	s.HandleFunc(`/{object:[a-zA-Z0-9_/\.]+}`, DeleteObjectHandler).Methods("DELETE").AppendFilterChain(filterChange)
	http.Handle("/", s)

	http.ListenAndServe(cfg.Server.Host+":"+cfg.Server.Port, s)
}
