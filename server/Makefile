# parameters
GO_BUILD=CGO_ENABLED=0 go build
GO_CLEAN=go clean
PROTOC=protoc
SERVER_BINARY_NAME=streamer

PROTO_GENERATED=storage/storage.pb.go

.DEFAULT_GOAL := build

%.pb.go: %.proto
	$(PROTOC) -I=storage/ --go_opt=paths=source_relative --go_out=storage/ $<

build: $(PROTO_GENERATED)
	$(GO_BUILD) -ldflags="$(LDFLAGS)" -o /bin/$(SERVER_BINARY_NAME) -v .
.PHONY: build

clean:
	$(GO_CLEAN)
	rm -f $(PROTO_GENERATED)
.PHONY: clean

all: build
.PHONY: all
