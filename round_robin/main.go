package main

import (
	"net/http/httputil",
	"net/url",
	"sync"
)

func main() {

}

type Backend struct {
	URL          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex
	//connections  int 
}

type LoadBalancer struct {
	backends []*Backend
	current  int
	mux      sync.Mutex
}
