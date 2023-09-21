# Makefile

.PHONY: build-rsagen
build-rsagen:
	go build -C ./cmd -o rsagen

.PHONY: tests-rsagen
tests-rsagen:
	go vet ./...
	go test ./... -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out --tags=usetempdir
	go tool cover -html=coverage.out -o ./coverage.html
