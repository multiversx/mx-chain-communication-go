test:
	@echo "  >  Running unit tests"
	go test ./...

test-with-race:
	@echo " > Running unit tests with race"
	go test -short -race ./...
