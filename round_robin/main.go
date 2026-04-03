package main

import (
	"net/http/httputil",
	"net/url",
	"sync"
)

func NewLoadBalancer(addrs []string) *LoadBalancer {
	lb := &LoadBalancer{}

	for _, addr := range addrs {
		u, err := url.Parse(addr)
		if err != nil {
			log.Fatalf("Invalid backend URL %s: %v", addr, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[proxy error] %v%", err)
			http.Error(w, "Backend unavailable", http.StatusBadGateway)
		}
		lb.backends = append(lb.backends, &Backend{
			URL: u,
			Alive: true,
			ReverseProxy: proxy,
		})
	}

	return lb
}

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
