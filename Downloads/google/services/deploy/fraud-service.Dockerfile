# syntax=docker/dockerfile:1.7
# Multi-stage build for FinGuard Go services.
#
# Build:  docker build -f deploy/fraud-service.Dockerfile -t finguard/fraud-service:dev .
# Run:    docker run --rm -p 8080:8080 -e MOCK_MODE=true finguard/fraud-service:dev

ARG SERVICE=fraud-service
ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/${SERVICE} ./cmd/${SERVICE}

FROM gcr.io/distroless/static:nonroot
ARG SERVICE
ENV SERVICE=${SERVICE}
COPY --from=build /out/${SERVICE} /${SERVICE}
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/fraud-service"]
