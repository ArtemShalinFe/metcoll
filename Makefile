# Makefile
tests: go-tests build-agent build-server ya-tests

build-agent:
	go build -buildvcs=false -C ./cmd/agent -o agent

build-server:
	go build -buildvcs=false -C ./cmd/server -o server

go-tests: go-mockgen
	go vet ./...
	go test ./... -v -coverpkg=./... -coverprofile=coverage.out 
	go tool cover -html=coverage.out -o ./coverage.html 

go-mockgen:
	mockgen -source=cmd/agent/main.go -destination=cmd/agent/mock_metcoll_client.go -package main

cur-test: build-agent build-server
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration13\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

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
	rm /tmp/test-db.json