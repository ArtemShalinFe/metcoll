# Makefile
tests: go-tests build-agent build-server ya-tests

build-agent:
	go build -buildvcs=false -C ./cmd/agent -o agent

build-server:
	go build -buildvcs=false -C ./cmd/server -o server

go-tests:
	go vet ./...
	go test ./... -v -coverpkg=./... -coverprofile=coverage.out 
	go tool cover -html=coverage.out -o ./coverage.html 

cur-test:
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration9\$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=8081 -file-storage-path=/tmp/test-db.json

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