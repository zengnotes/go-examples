package main

import (
	"os"
	"log"
	"strconv"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"io"
)

var transport = &http.Transport{DisableKeepAlives: false, DisableCompression: false}

func getConfigFromEnvOrDieTrying()(endpoint string, port int) {
	endpoint = os.Getenv("endpoint")
	if endpoint == "" {
		log.Fatal("No endpoint specified")
	}

	port,err := strconv.Atoi(os.Getenv("port"))
	if err != nil {
		log.Fatal("No valid port specified")
	}

	return
}

func makeHandler(endpoint string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = endpoint
		req.Host = endpoint
		resp, err := transport.RoundTrip(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(rw, resp.Body)
		resp.Body.Close()
	}
}


func main() {
	endpoint, port := getConfigFromEnvOrDieTrying()
	log.Println("Port:", port, "Endpoint:",endpoint)

	r := mux.NewRouter()
	r.HandleFunc("/",makeHandler(endpoint))
	http.Handle("/",r)
	err := http.ListenAndServe(fmt.Sprintf(":%d",port), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}