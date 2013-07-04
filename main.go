package main

import (
	// "bitbucket.org/taruti/mimemagic"
	"crypto/md5"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/peterbourgon/diskv"
	"io"
	"io/ioutil"
	"net/http"
	// "os"
	"strings"
)

// holds Diskv database
var store *diskv.Diskv

func BucketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func GETObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// bucket := vars["bucket"]
	object := vars["object"]

	md5 := md5sum(object)

	val, err := store.Read(md5)

	if err != nil {
		// fmt.Printf("%s", err)
		// panic(fmt.Sprintf("key %s had no value", object))
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 Not Found")
		return
	}

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

	// w.Header().Set("Content-Type", content_type)

	fmt.Fprintf(w, "%s", val)
}

func POSTObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	fmt.Fprintf(w, "POST %s!\n", r.URL.Path[1:])
	fmt.Fprintf(w, "bucket: %s\n", bucket)
	fmt.Fprintf(w, "object: %s\n", object)

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	content_type := handler.Header.Get("Content-Type")

	md5 := md5sum(object)

	fmt.Printf("Uploaded: %s to %s with Content-Type: %s\n", handler.Filename,
		strings.Join(BlockTransform(md5), "/")+
			"/"+
			md5,
		content_type)

	store.Write(md5, []byte(data))
}

func DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	fmt.Fprintf(w, "DELETE %s!\n", r.URL.Path[1:])
	fmt.Fprintf(w, "bucket: %s\n", bucket)
	fmt.Fprintf(w, "object: %s\n", object)

	md5 := md5sum(object)

	fmt.Printf("Deleted: %s at %s\n", object,
		strings.Join(BlockTransform(md5), "/")+
			"/"+
			md5)

	// Erase the key+value from the store (and the disk).
	store.Erase(md5)
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

func main() {

	// Initialize a new diskv store, rooted at "store", with a 1MB cache.
	store = diskv.New(diskv.Options{
		BasePath:     "store",
		Transform:    BlockTransform,
		CacheSizeMax: 1024 * 1024,
	})

	r := mux.NewRouter()
	r.HandleFunc("/{bucket}", BucketHandler)
	r.HandleFunc(`/{bucket}/{object:[a-zA-Z_/\.]+}`, GETObjectHandler).Methods("GET")
	r.HandleFunc(`/{bucket}/{object:[a-zA-Z_/\.]+}`, POSTObjectHandler).Methods("POST")
	r.HandleFunc(`/{bucket}/{object:[a-zA-Z_/\.]+}`, DeleteObjectHandler).Methods("DELETE")
	http.Handle("/", r)

	http.ListenAndServe(":8080", r)
}
