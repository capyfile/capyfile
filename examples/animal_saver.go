package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

const (
	Byte = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
	PiByte
	EiByte
)

// This is a simple app that proxies requests to capyfile server that validates, transforms,
// and saves files.
//
// Here you can also see one way of providing parameters to capyfile server.
func main() {
	http.HandleFunc("/", capyfileProxyHandler)

	http.ListenAndServe(":8080", nil)
}

func capyfileProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := strings.Split(
		strings.TrimLeft(r.URL.Path, "/"),
		"/")

	if len(path) != 1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	animal := path[0]

	if animal == "seagull" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Different animals have different requirements for max file size and file type.
	if animal == "capybara" {
		r.Header.Set(
			"X-Capyfile-FileSizeValidate-MaxFileSize",
			strconv.Itoa(100*MiByte))
		r.Header.Set(
			"X-Capyfile-FileTypeValidate-AllowedMimeTypes",
			`["image/jpeg", "image/png", "image/gif", "image/bmp", "image/webp", "image/heic", "image/heif"]`)
	} else {
		r.Header.Set(
			"X-Capyfile-FileSizeValidate-MaxFileSize",
			strconv.Itoa(1*MiByte))
		r.Header.Set(
			"X-Capyfile-FileTypeValidate-AllowedMimeTypes",
			`["image/jpeg", "image/png"]`)
	}
	// Every animal is going to be stored in their own bucket.
	// We assume that all the necessary buckets are already created.
	r.Header.Set("X-Capyfile-S3Upload-Bucket", animal)

	r.URL.Path = "/upload/animal"

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "localhost:8024",
	})
	proxy.ServeHTTP(w, r)
}
