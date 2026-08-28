FROM golang:1.27-alpine AS build

ARG GOOSE_VERSION=v3.27.3

RUN CGO_ENABLED=0 GOBIN=/out go install \
    github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /migrations

COPY --from=build /out/goose /usr/local/bin/goose
COPY migrations/*.sql ./

ENTRYPOINT ["/usr/local/bin/goose"]
