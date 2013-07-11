package main

import (
	"crypto/md5"
	"io"
	"fmt"
	"strings"
)

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

// md5sum method

func md5sum(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}


func ActualPaths(s string) string {
	md5 := md5sum(s)
	return strings.Join(BlockTransform(md5), "/") +
		"/" +
		md5
}