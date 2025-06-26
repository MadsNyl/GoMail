run:
	@go run *.go

.PHONY: worker
worker:
	@go run ./worker/main.go

build-redis:
	@docker build -t redis:latest -f Dockerfile.redis .

redis: build-redis
	@docker run -d --name redis -p 6379:6379 redis:latest

stop-redis:
	@docker stop redis || true
	@docker rm redis || true
