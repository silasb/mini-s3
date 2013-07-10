package main

import (
	// "bitbucket.org/taruti/mimemagic"
	"crypto/md5"
	"fmt"
	"github.com/studygolang/mux"
	"github.com/peterbourgon/diskv"
	"io"
	"io/ioutil"
	"net/http"
	// "os"
	"code.google.com/p/gcfg"
	"strings"
	"filter"
	"github.com/jmhodges/levigo"
)



// holds Diskv database
var store *diskv.Diskv

var meta_store *levigo.DB

func BucketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "chop word, carry water")
}

func GETObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["subdomain"]
	object := vars["object"]

	path := bucket + "/" + object

	md5 := md5sum(path)

	val, err := store.Read(md5)

	// check if we have clients that are getting sending ETag information
	etags := r.Header["If-None-Match"]
	if len(etags) != 0 {
		if etags[0] == md5sum(object) {
			w.WriteHeader(http.StatusNotModified)
			return
		}		
	}

	// since we didn't get a hit on the ETag, lets write it out.
	w.Header().Set("ETag", md5sum(object))

	if err != nil {
		// fmt.Printf("%s", err)
		// panic(fmt.Sprintf("key %s had no value", object))
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 Not Found\n")
		return
	}

	fmt.Printf("GET: %s at %s\n", object, ActualPaths(path))

	ro := levigo.NewReadOptions()
	data, err := meta_store.Get(ro, []byte(object+"-content-type"))
	defer ro.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	content_type := string(data)

	// fmt.Println(string(data))

	// Determine content-type manually

	// b := make([]byte, 1024)
	// file, err := os.Open("store/" + strings.Join(BlockTransform(md5), "/") +
	// 	"/" +
	// 	md5)
	// if err != nil {
	// 	panic(err)
	// }
	// file.Read(b)
	// defer file.Close()

	// content_type := mimemagic.Match("", b)

	w.Header().Set("Content-Type", content_type)

	fmt.Fprintf(w, "%s", val)
}

func POSTObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["subdomain"]
	object := vars["object"]

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprint(w, "406 Form field \"file\" not present\n")
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintf(w, "406 %s", err)
		return
	}

	fmt.Fprintf(w, "bucket: %s\n", bucket)
	fmt.Fprintf(w, "object: %s\n", object)

	content_type := handler.Header.Get("Content-Type")

	wo := levigo.NewWriteOptions()
	err = meta_store.Put(wo, []byte(object+"-content-type"), []byte(content_type))
	defer wo.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	path := bucket + "/" + object
	md5 := md5sum(path)

	fmt.Printf("Uploaded: %s to %s with Content-Type: %s\n", handler.Filename,
	ActualPaths(path),
	content_type)

	store.Write(md5, []byte(data))
}

func DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["subdomain"]
	object := vars["object"]

	fmt.Fprintf(w, "bucket: %s\n", bucket)
	fmt.Fprintf(w, "object: %s\n", object)

	path := bucket + "/" + object
	md5 := md5sum(path)

	fmt.Printf("Deleted: %s at %s\n", object, ActualPaths(path))

	// Erase the key+value from the store (and the disk).
	store.Erase(md5)
}

func ActualPaths(s string) string {
	md5 := md5sum(s)
	return strings.Join(BlockTransform(md5), "/") +
		"/" +
		md5
}

// transform methods

const (
	transformBlockSize = 6 // grouping of chars per directory depth
)

func BlockTransform(s string) []string {
	sliceSize := len(s) / transformBlockSize
	pathSlice := make([]string, sliceSize)
	for i := 0; i < sliceSize; i++ {
		from, to := i*transformBlockSize, (i*transformBlockSize)+transformBlockSize
		pathSlice[i] = s[from:to]
	}
	return pathSlice
}

func md5sum(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// end transform methods

type Config struct {
	Server struct {
		Host  string
		Port  string
		DomainName string
		Store string
	}
}

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

	// Initialize a new diskv store, rooted at "store", with a 1MB cache.
	store = diskv.New(diskv.Options{
		BasePath:     cfg.Server.Store,
		Transform:    BlockTransform,
		CacheSizeMax: 1024 * 1024,
	})

	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(3<<30))
	opts.SetCreateIfMissing(true)
	meta_store, err = levigo.Open("./meta_store", opts)

	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(meta_store)

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
