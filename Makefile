# Makefile
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: tests
tests: go-tests build-agent build-server ya-tests

.PHONY: build-agent
build-agent:
	go build -buildvcs=false -C ./cmd/agent -o agent

.PHONY: build-server
build-server:
	go build -buildvcs=false -C ./cmd/server -o server

.PHONY: go-tests
go-tests:
	go vet ./...
	go test ./... -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out 
	go tool cover -html=coverage.out -o ./coverage.html

.PHONY: cur-test
cur-test: build-agent build-server
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration14\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -key="olala/poslednieTesti" -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: ya-tests
ya-tests:	
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration1\$$ -source-path=. -binary-path=cmd/server/server
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration2[AB]*\$$ -source-path=. -agent-binary-path=cmd/agent/agent
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration3[AB]*\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration4\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration5\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration6\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration7\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration8\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration9\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -file-storage-path=/tmp/test-db.json
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration10[AB]\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration11\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration12\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration13\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration14\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -key="olala/poslednieTesti" -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'
	rm /tmp/test-db.json

.PHONY: lint
lint:
	[ -d $(ROOT_DIR)/golangci-lint ] || mkdir -p $(ROOT_DIR)/golangci-lint
	docker run --rm \
    -v $(ROOT_DIR):/app \
    -v $(ROOT_DIR)/golangci-lint/.cache:/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.53.3 \
        golangci-lint run \
        -c .golangci-lint.yml \
    > ./golangci-lint/report.json