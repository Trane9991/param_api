package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Response struct {
	status  int
	Message string
	Data    interface{}
}

var (
	CACHE      = make(map[string]Response)
	region     = os.Getenv("AWS_REGION")
	AppVersion = "0.0.1"
)

func main() {
	api()
}

func api() {
	router := mux.NewRouter().StrictSlash(true)
	registerHandlers(router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Println("Validating Config") //todo, validate config
	if region == "" {
		log.Fatal("Environment variable AWS_REGION undefined")
		//todo, check against list of known regions
	}
	log.Println("Started: Ready to serve")
	log.Fatal(http.ListenAndServe(":8080", loggedRouter)) //todo, refactor to make port dynamic
}

func registerHandlers(r *mux.Router) {
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.HandleFunc("/params", paramsHandler).Methods("POST")
	r.HandleFunc("/file", fileHandler).Methods("POST")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m = make(map[string]string)
	m["error"] = fmt.Sprintf("Route %s not found with method %s, please check request and try again",
		r.URL.Path, r.Method)
	resp := Response{Data: m, status: http.StatusNotFound}
	JSONResponseHandler(w, resp)
}

func JSONResponseHandler(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)
	json.NewEncoder(w).Encode(resp.Data)
}
