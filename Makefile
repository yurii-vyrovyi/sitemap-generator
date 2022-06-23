
PHONY: deps
deps:
	go mod tidy

PHONY: lint
lint:
	golangci-lint run --allow-parallel-runners -v -c .golangci.yml


.PHONY: unit_test
unit_test:
	go test \
		-race \
		-count=1 \
		-v \
		./...

.PHONY: mockgen
mockgen:
	GO111MODULE=off go install github.com/golang/mock/mockgen
	go generate ./...