FROM golang:1.21.1-alpine3.18 AS build

LABEL site="ystv-streamer-recorder"

ARG STREAMER_VERSION_ARG

VOLUME /logs
VOLUME /tank/pending_edits/

WORKDIR /src/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Set build variables
RUN echo -n "-X 'main.Version=$STREAMER_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/recorder

EXPOSE 1323

ENTRYPOINT ["/bin/recorder"]