SRC := *.go
BIN := notifierd

run:
	@go run $(SRC)
build:
	@go build -o $(BIN) $(SRC)