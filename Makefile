include app.env

redis:
	docker run --name redis3 -p 6379:6379 -d redis:6.2-alpine3.13 redis-server --requirepass ${REDIS_PASS}

test:
	@sh ./test.sh
	@echo "================================================" | GREP_COLOR='01;33' grep -E --color '^.*=.*'
	@printf "\033[33mCoverage\033[0m"
	@echo ""
	@sh ./cover.sh

lint:
	golangci-lint run

.PHONY: redis test lint