# See https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:1.15-alpine AS builder

# Git is used for dependencies
RUN apk update \
    && apk add \
        --no-cache \
        git \
        ca-certificates

ENV USER=gonyexpress
ENV UID=1000
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    -u "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/gonyexpress/

COPY go.mod go.sum ./

RUN go mod download \
    && go mod verify

COPY . .

ARG TARGET

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o /go/bin/gonyexpress \
    -ldflags="-w -s" \
    bin/${TARGET}/${TARGET}.go

FROM scratch

COPY --from=builder \
    /etc/passwd \
    /etc/group \
    /etc/
COPY --from=builder \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/
COPY --from=builder \
    /go/bin/gonyexpress \
    /go/bin/gonyexpress

USER gonyexpress:gonyexpress

ENTRYPOINT [ "/go/bin/gonyexpress", "--rabbitmq", "amqp://guest:guest@rabbitmq:5672/"]
