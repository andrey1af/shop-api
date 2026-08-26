FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN test "$(go list -f '{{.Name}}' .)" = "main" \
    && CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/api-service \
        .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/api-service /usr/local/bin/api-service

EXPOSE 8090

ENTRYPOINT ["/usr/local/bin/api-service"]
