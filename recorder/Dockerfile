FROM golang:1.21.7-alpine3.19 AS build

LABEL site="ystv-streamer-recorder"

ARG STREAMER_VERSION_ARG
ARG STREAMER_RECORDER_USER_UID

# Create recorder.
ENV USER=recorder
ENV UID=$STREAMER_RECORDER_USER_UID

# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

VOLUME /logs
VOLUME /tank/pending_edits/

WORKDIR /src/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN apk update && apk upgrade && \
    apk add --no-cache ffmpeg ca-certificates && \
    update-ca-certificates

# Set build variables
RUN echo -n "-X 'main.Version=$STREAMER_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/recorder

EXPOSE 1323

# Use an unprivileged user.
USER recorder:recorder

ENTRYPOINT ["/bin/recorder"]