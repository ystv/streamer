FROM golang:1.21.1-alpine3.18 AS build

LABEL site="ystv-streamer-forwarder"

VOLUME /logs

WORKDIR /src/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/forwarder

EXPOSE 1323

ENTRYPOINT ["/bin/forwarder"]