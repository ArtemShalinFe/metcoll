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

.PHONY: go-tests-with-tags
go-tests-with-tags:
	go vet ./...
	go test ./... --tags=usetempdir -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out -covermode=atomic 
	go tool cover -html=coverage.out -o ./coverage.html


.PHONY: unit-tests
unit-tests:
	go vet ./...
	go test ./... -v -race -count=1 -coverpkg=./... -coverprofile=unit_coverage.out --tags=usetempdir
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
mocks: protoc
	mockgen -source=internal/metrics/metrics.go -destination=internal/metrics/mock_metrics.go -package metrics
	mockgen -source=internal/metcoll/handlers.go -destination=internal/metcoll/mock_handlers.go -package metcoll
	mockgen -source=internal/metcoll/metcoll_grpc.pb.go -destination=internal/metcoll/mock_grpc_pb.go -package metcoll

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
	[ -d $(ROOT_DIR)/keys ] || mkdir -p $(ROOT_DIR)/keys
	$(MAKE) -C $(ROOT_DIR)/rsagen build-rsagen
	mv $(ROOT_DIR)/rsagen/cmd/rsagen $(ROOT_DIR)/keys
	$(ROOT_DIR)/keys/rsagen -o "$(ROOT_DIR)/keys" -b "10000"
	rm -f $(ROOT_DIR)/keys/rsagen

.PHONY: protoc
protoc:
	protoc proto/v1/*.proto  --proto_path=proto/v1 \
	--go_out=internal/metcoll --go_opt=module=github.com/ArtemShalinFe/metcoll/internal/metcoll \
	--go-grpc_out=internal/metcoll --go-grpc_opt=module=github.com/ArtemShalinFe/metcoll/internal/metcoll
	