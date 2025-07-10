FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod ./
COPY go.su[m] ./

RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /app/build/openbpl \
    ./cmd/server

FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1001 -S openbpl && \
    adduser -u 1001 -S openbpl -G openbpl

WORKDIR /app

COPY --from=builder /app/build/openbpl /app/openbpl

COPY --from=builder /app/static /app/static

RUN mkdir -p /app/logs && \
    chown -R openbpl:openbpl /app

USER openbpl

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/openbpl"]

FROM golang:1.24-alpine AS development

RUN apk add --no-cache git ca-certificates tzdata curl

RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod ./
COPY go.su[m] ./

RUN go mod download

COPY .air.toml .air.toml

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

FROM builder AS testing

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

RUN go test -v ./...

RUN golangci-lint run