
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