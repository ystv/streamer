FROM golang:1.21.1-alpine3.18 AS build

LABEL stage="builder"

VOLUME /db

WORKDIR /src/

COPY . .

RUN go mod download

RUN apk update && apk add git && apk add make && apk add protoc && apk add protoc-gen-go --repository https://dl-cdn.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

RUN GOOS=linux GOARCH=amd64 make

COPY ./streamer /bin

EXPOSE 8084
ENTRYPOINT ["/bin/streamer"]