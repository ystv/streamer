# parameters
GO_BUILD=CGO_ENABLED=0 go build
GO_CLEAN=go clean
PROTOC=protoc
SERVER_BINARY_NAME=streamer
SERVER_FOLDER=server
RECORDER_BINARY_NAME=recorder
RECORDER_FOLDER=recorder
FORWARDER_BINARY_NAME=forwarder
FORWARDER_FOLDER=forwarder

PROTO_GENERATED=server/storage/storage.pb.go

.DEFAULT_GOAL := build

%.pb.go: %.proto
	$(PROTOC) -I=server/storage/ --go_opt=paths=source_relative --go_out=server/storage/ $<

build: $(PROTO_GENERATED)
	mkdir -p bin
	cd server && $(GO_BUILD) -o $(SERVER_BINARY_NAME) -v . && cd ..
	cp server/$(SERVER_BINARY_NAME) bin
	cd forwarder && $(GO_BUILD) -o $(FORWARDER_BINARY_NAME) -v . && cd ..
	cp forwarder/$(FORWARDER_BINARY_NAME) bin
	cd recorder && $(GO_BUILD) -o $(RECORDER_BINARY_NAME) -v . && cd ..
	cp recorder/$(RECORDER_BINARY_NAME) bin
.PHONY: build

clean:
	$(GO_CLEAN)
	rm -f $(PROTO_GENERATED)
.PHONY: clean

all: build
.PHONY: all
