package main

import (
	"io/ioutil"
	"net/http"
	"fmt"
	"path"
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

	bucket_path := bucket + "/" + object

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

	fmt.Printf("GET: %s at %s\n", object, ActualPaths(bucket_path))

	// get any meta information for the object
	ro := levigo.NewReadOptions()
	defer ro.Close()

	// only meta information is content-type for now
	data, err := meta_store.Get(ro, []byte(object+"-content-type"))
	if err != nil {
		fmt.Println(err)
		return
	}

	content_type := string(data)

	w.Header().Set("Content-Type", content_type)

	http.ServeFile(w, r, path.Join(abs_store_path, ActualPaths(bucket_path)))
}

// POST /:object
//
func POSTObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["subdomain"]
	object := vars["object"]

	// open the file that got uploaded
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprint(w, "406 Form field \"file\" not present\n")
		return
	}
	defer file.Close()

	// read the file into the data structure.
	// TODO: don't read it into memory, but io.Copy it from current location to store path?
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintf(w, "406 %s", err)
		return
	}

	// get content type from the file that we uploaded
	content_type := handler.Header.Get("Content-Type")

	// open the meta-store to write the content-type of the file
	wo := levigo.NewWriteOptions()
	defer wo.Close()

	// write the content type to the meta-store
	err = meta_store.Put(wo, []byte(object+"-content-type"), []byte(content_type))
	if err != nil {
		fmt.Println(err)
		return
	}

	bucket_path := bucket + "/" + object
	md5 := md5sum(bucket_path)

	fmt.Printf("Uploaded: %s to %s with Content-Type: %s\n",
		handler.Filename,
		ActualPaths(bucket_path),
		content_type,
	)

	// write data to data-store
	store.Write(md5, []byte(data))

	// write response to client
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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
