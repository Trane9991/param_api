package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type fileRequest struct {
	Path    string
	Version string
}

func (f fileRequest) valid() bool {
	if f.Path == "" || f.Version == "" {
		return false
	}
	return true
}

func (p fileRequest) badRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	var m = make(map[string]string)
	expected := strings.ToLower(fmt.Sprintf("%+v", fileRequest{"STRING", "STRING"}))
	m["error"] = fmt.Sprintf("Bad request, expected: %s, got: %s", expected, strings.ToLower(fmt.Sprintf("%+v", p)))
	resp := Response{Data: m, status: http.StatusBadRequest}
	JSONResponseHandler(w, resp)
}

func (f fileRequest) identifier() string {
	return fmt.Sprintf("%s@%s", f.Path, f.Version)
}

func (f fileRequest) cacheKey() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(f.identifier())))
}

func parseFileRequestBody(b io.ReadCloser) fileRequest {
	decoder := json.NewDecoder(b)
	var f fileRequest
	err := decoder.Decode(&f)
	if err != nil {
		log.Printf("encountered issue decoding request body; %s", err.Error())
		return fileRequest{}
	}
	return f
}

func (f fileRequest) getData() string {
	c := ssmClient{NewClient(region)}

	return c.GetKey(f.Path, f.Version)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	f := parseFileRequestBody(r.Body)
	if !f.valid() {
		f.badRequest(w)
		return
	}

	log.Printf("Processing request for %s uniquely identified as %+v", f.identifier(), f.cacheKey())
	cached, ok := CACHE[f.cacheKey()]
	if ok {
		log.Printf("Retrieved parameters from cache")
		JSONResponseHandler(w, cached)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(f.getData()))
}
