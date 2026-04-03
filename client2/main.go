package main

import (
	"fmt"
	"net/http"
)

func statusCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Client 2 Server is healthy")
}

func headers(w http.ResponseWriter, r *http.Request) {
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	http.HandleFunc("/health", statusCheckHandler)
	http.HandleFunc("/headers", headers)

	fmt.Print("Client 2 server is running")
	http.ListenAndServe(":8083", nil)
}
