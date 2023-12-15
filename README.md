# golang-server

Minimal golang server

# Start in docker

1. Build image: `docker build -t golang-server .`
2. Run application: `docker run -d -p 8080:8080 golang-server go run cmd/app/main.go`
3. Open in browser: `http://localhost:8080/hello`
4. Stop application: `docker ps` and `docker stop <container_id>`