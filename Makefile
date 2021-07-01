include app.env

redis:
	docker run --name redis3 -p 6379:6379 -d redis:6.2-alpine3.13 redis-server --requirepass ${REDIS_PASS}

test:
	go test ./...  -v -coverprofile cover.out
	@echo "================================================"
	@echo "Coverage"
	go tool cover -func cover.out

lint:
	golangci-lint run

.PHONY: redis test lint