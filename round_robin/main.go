package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

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

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

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
			URL:          u,
			Alive:        false,
			ReverseProxy: proxy,
		})
	}

	return lb
}

func (lb *LoadBalancer) NextBackend() *Backend {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	total := len(lb.backends)
	for i := 0; i < total; i++ {
		idx := (lb.current + 1) % total

		if lb.backends[idx].IsAlive() {
			lb.current = (idx + 1) % total
			return lb.backends[idx]
		}
	}

	return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.NextBackend()
	if backend == nil {
		http.Error(w, "There are no healthy backends available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("[lb] -> forwarding to %s", backend.URL)
	backend.ReverseProxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) StartHealthChecks(interval time.Duration) {
	for _, b := range lb.backends {
		go func(backend *Backend) {
			for range time.Tick(interval) {
				conn, err := net.DialTimeout("tcp", backend.URL.Host, 2*time.Second)
				if err != nil {
					log.Printf("[health] backend %s is DOWN", backend.URL)
					backend.SetAlive(false)
				} else {
					conn.Close()
					log.Printf("[health] backend %s is UP", backend.URL)
					backend.SetAlive(true)
				}
			}
		}(b)
	}
}

func main() {
	lb := NewLoadBalancer([]string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	})

	lb.StartHealthChecks(5 * time.Second)

	log.Printf("[lb] Load balancer is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", lb))
}
