obu:
	@go build -o bin/obu cmd/obu/main.go
	@./bin/obu

receiver:
	@go build -o bin/receiver cmd/data_receiver/main.go
	@./bin/receiver


.PHONY: obu 