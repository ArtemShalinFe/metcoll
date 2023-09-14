# Makefile
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

VERSION=$(shell cat VERSION)
DATETIME=$(shell date +'%Y/%m/%d %H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags="-X github.com/ArtemShalinFe/metcoll/internal/build.buildVersion=$(VERSION) -X 'github.com/ArtemShalinFe/metcoll/internal/build.buildDate=$(DATETIME)' -X github.com/ArtemShalinFe/metcoll/internal/build.buildCommit=$(COMMIT)"

.PHONY: tests
tests: go-tests build-agent build-server ya-tests

.PHONY: build-agent
build-agent:
	go build \
		-C ./cmd/agent \
		-o agent \
		$(LDFLAGS)
		

.PHONY: build-server
build-server:
	go build \
		-C ./cmd/server \
		-o server \
		$(LDFLAGS)

.PHONY: build-staticlint
build-staticlint:
	go build \
		-C ./cmd/staticlint \
		-o staticlint

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

.PHONY: mocks
mocks:
	mockgen -source=internal/metrics/metrics.go -destination=internal/metrics/mock_metrics.go -package metrics

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

.PHONY: cryptokeys
cryptokeys:
	openssl req -x509 -nodes -days 365 -newkey rsa:16384 -keyout $(ROOT_DIR)/keys/private.pem -out $(ROOT_DIR)/keys/cert.pem
	openssl rsa -in $(ROOT_DIR)/keys/private.pem -pubout -out $(ROOT_DIR)/keys/public.pem