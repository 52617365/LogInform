check:
	# Format and organize
	goimports -w .
	gofmt -w .

	go mod tidy

	# Security
	govulncheck ./...

	# Static analysis
	go vet ./...
	deadcode .
	staticcheck ./...
	golangci-lint fmt
	golangci-lint run --fix

	# Testing
	go test -race ./...

run:
	go run .

doctor:
	make check
	docker build -t loginform .
	docker run --rm loginform

poll:
	find . -name "*.go" | entr -r make check