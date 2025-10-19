# requestjar-go

# Testing

## Running tests

Locally you can run all tests with:

```sh
make test
```

Or directly with Go:

```sh
go test ./...
```

## Coverage

Generate a coverage report locally with:

```sh
make coverage
```

This will create a `coverage.out` file and print a brief summary. You can view the full HTML report with:

```sh
go tool cover -html=coverage.out
```
