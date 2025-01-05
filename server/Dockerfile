FROM golang:1.23.4-alpine3.21 AS build

LABEL stage="builder"

ARG STREAMER_VERSION_ARG

VOLUME /db

WORKDIR /src/

COPY . .

RUN go mod download

RUN apk update && apk add git make protoc && apk add protoc-gen-go --repository https://dl-cdn.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

# Set build variables
RUN echo -n "-X 'main.Version=$STREAMER_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

RUN GOOS=linux GOARCH=amd64 make LDFLAGS="$(cat ./ldflags)"

EXPOSE 8084

ENTRYPOINT ["/bin/streamer"]