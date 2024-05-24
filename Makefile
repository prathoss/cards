.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

.PHONY: cover
cover:
	go test -cover ./...

.PHONY: coverage-report
coverage-report:
	go test -coverprofile coverage.out
	go tool cover -html=coverage.out
