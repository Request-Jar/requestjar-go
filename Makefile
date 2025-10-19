.PHONY: lint test coverage

lint:
	golangci-lint run --no-config --enable=govet --enable=staticcheck --enable=errcheck --enable=loggercheck

test:
	@echo "Running Go tests..."
	go test ./...

coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | sed -n '1,3p'
	@echo "Full coverage report written to coverage.out"
