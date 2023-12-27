FROM golang:1.21.5-alpine3.19 AS build

LABEL site="ystv-streamer-forwarder"

ARG STREAMER_VERSION_ARG

VOLUME /logs

WORKDIR /src/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Set build variables
RUN echo -n "-X 'main.Version=$STREAMER_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/forwarder

FROM alpine

LABEL site="ystv-streamer-forwarder"

RUN apk update && apk upgrade && \
    apk add --no-cache ffmpeg ca-certificates

FROM scratch

LABEL site="ystv-streamer-forwarder"
# Copy binary
COPY --from=build /bin/forwarder /bin/forwarder
EXPOSE 1323
ENTRYPOINT ["/bin/forwarder"]