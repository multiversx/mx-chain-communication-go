test:
	@echo "  >  Running unit tests"
	go test ./...

test-short-race:
	@echo " > Running short unit tests in race mode"
	go test -short -race ./...
