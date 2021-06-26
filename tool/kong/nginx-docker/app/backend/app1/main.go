package main

import (
	"fmt"
	"net/http"
)

func main() {
	s := http.Server{
		Addr:    "localhost:8080",
		Handler: handler{},
	}
	s.ListenAndServe()
}

type handler struct{}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Print("\n\n")
	fmt.Printf("host: %v\n", r.Host)
	for k, v := range r.Header {
		fmt.Printf("%v: %v\n", k, v)
	}
}
