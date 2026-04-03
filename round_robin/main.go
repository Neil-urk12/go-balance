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

type Server struct {
	URL          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex
	//connections  int
}

type LoadBalancer struct {
	servers []*Server
	current int
	mux     sync.Mutex
}

func (s *Server) SetAlive(alive bool) {
	s.mux.Lock()
	s.Alive = alive
	s.mux.Unlock()
}

func (s *Server) IsAlive() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.Alive
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
		lb.servers = append(lb.servers, &Server{
			URL:          u,
			Alive:        false,
			ReverseProxy: proxy,
		})
	}

	return lb
}

func (lb *LoadBalancer) NextServer() *Server {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	total := len(lb.servers)
	for i := 0; i < total; i++ {
		idx := (lb.current + 1) % total

		if lb.servers[idx].IsAlive() {
			lb.current = (idx + 1) % total
			return lb.servers[idx]
		}
	}

	return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := lb.NextServer()
	if server == nil {
		http.Error(w, "There are no healthy backends available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("[lb] -> forwarding to %s", server.URL)
	server.ReverseProxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) StartHealthChecks(interval time.Duration) {
	for _, s := range lb.servers {
		go func(server *Server) {
			for range time.Tick(interval) {
				conn, err := net.DialTimeout("tcp", server.URL.Host, 2*time.Second)
				if err != nil {
					log.Printf("[health] backend %s is DOWN", server.URL)
					server.SetAlive(false)
				} else {
					conn.Close()
					log.Printf("[health] backend %s is UP", server.URL)
					server.SetAlive(true)
				}
			}
		}(s)
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
