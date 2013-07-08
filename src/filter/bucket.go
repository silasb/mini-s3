package filter

import (
	"github.com/studygolang/mux"
	"net/http"
	"fmt"
)

type BucketFilter struct {
	*mux.EmptyFilter
}

func (this *BucketFilter) PreFilter(rw http.ResponseWriter, req *http.Request) bool {
	vars := mux.Vars(req)
	bucket := vars["subdomain"]

	if bucket == "" {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprint(rw, "404 No Bucket Provided\n")
		return false
	}

	return true
}