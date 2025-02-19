
all: run

bin/app: main/main.go ./proto/sample.pb.go proto/unmarshal.go
	mkdir -p bin
	go build -o ./bin/app main/main.go

./proto/sample.pb.go: ./sample/sample.proto
	protoc -I=./sample --go_out=./ ./sample/sample.proto

run: bin/app
	./bin/app

test: main
	go test ./...

.PHONY: run test
