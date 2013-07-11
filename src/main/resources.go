package main

import (
	"io/ioutil"
	"net/http"
	"fmt"
	"github.com/studygolang/mux"
	"github.com/jmhodges/levigo"
	// "bitbucket.org/taruti/mimemagic"
)

// GET /
//
func BucketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "chop word, carry water")
}

// GET /:object
//
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
	defer ro.Close()

	data, err := meta_store.Get(ro, []byte(object+"-content-type"))
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

// POST /:object
//
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
	defer wo.Close()

	err = meta_store.Put(wo, []byte(object+"-content-type"), []byte(content_type))
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

// DELETE /:object
//
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

	wo := levigo.NewWriteOptions()
	defer wo.Close()

	err := meta_store.Delete(wo, []byte(object+"-content-type"))
	if err != nil {
		fmt.Println(err)
		return
	}
}
