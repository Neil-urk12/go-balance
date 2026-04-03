# go-balance

Tiny Go demo: a round-robin reverse-proxy load balancer with two backend HTTP services.

Structure
- round_robin/ — load balancer (listens on :8080)
- client1/ — backend server (listens on :8081)
- client2/ — backend server (listens on :8083)

Quick run
- Start backends: `go run ./client1` and `go run ./client2`
- Start balancer: `go run ./round_robin`

Build (per-service)
- `cd client1 && go build -o ./client1`
- `cd client2 && go build -o ./client2`
- `cd round_robin && go build -o ./round_robin`

Smoke tests
- `curl http://localhost:8080/headers -v -H "X-Test: 1"`
- `curl http://localhost:8081/health`

Notes
- Backends expose `/health` and `/headers`.
- Health checks are TCP-level; edit `round_robin/main.go` to change backends or intervals.
