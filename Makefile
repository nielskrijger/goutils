PKGS := $(shell go list ./...| grep -v /mocks)

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(PKGS) -v -coverprofile=coverage.out -timeout 10s

.PHONY: humantest
humantest:
	LOG_HUMAN=true richgo test -v -p=1 -timeout=60s $(PKGS)

.PHONY: coverage
coverage:
	go tool cover -html=coverage.out
