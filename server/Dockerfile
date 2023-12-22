FROM golang:1.21.5-alpine3.19 AS build

LABEL stage="builder"

ARG STREAMER_VERSION_ARG

VOLUME /db

WORKDIR /src/

COPY . .

RUN go mod download

RUN apk update && apk add git && apk add make && apk add protoc && apk add protoc-gen-go --repository https://dl-cdn.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

# Set build variables
RUN echo -n "-X 'main.Version=$STREAMER_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

RUN GOOS=linux GOARCH=amd64 make LDFLAGS="$(cat ./ldflags)"

EXPOSE 8084

ENTRYPOINT ["/bin/streamer"]