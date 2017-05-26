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

type paramRequest struct {
	Application string
	Environment string
	Version     string
	Landscape   string
}

func (p paramRequest) valid() bool {
	if p.Application == "" || p.Environment == "" || p.Version == "" || p.Landscape == "" {
		return false
	}
	return true
}

func (p paramRequest) badRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	var m = make(map[string]string)
	expected := strings.ToLower(fmt.Sprintf("%+v", paramRequest{"STRING", "STRING", "STRING", "STRING"}))
	m["error"] = fmt.Sprintf("Bad request, expected: %s, got: %s", expected, strings.ToLower(fmt.Sprintf("%+v", p)))
	resp := Response{Data: m, status: http.StatusBadRequest}
	JSONResponseHandler(w, resp)
}

func (p paramRequest) envPrefix() string {
	return fmt.Sprintf("%s.%s.%s", p.Landscape, p.Environment, p.Application)
}

func (p paramRequest) cacheKey() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(p.identifier())))
}

func (p paramRequest) identifier() string {
	return fmt.Sprintf("%s@%s", p.envPrefix(), p.Version)
}

func parseParamRequestBody(b io.ReadCloser) paramRequest {
	decoder := json.NewDecoder(b)
	var p paramRequest
	err := decoder.Decode(&p)
	if err != nil {
		log.Printf("encountered issue decoding request body; %s", err.Error())
		return paramRequest{}
	}
	return p
}

func (p paramRequest) getData() map[string]string {
	c := ssmClient{NewClient(region)}
	paramNames := c.WithPrefix(p.envPrefix())
	return paramNames.IncludeHistory(c).withVersion(p.Version) //todo, return error
}
