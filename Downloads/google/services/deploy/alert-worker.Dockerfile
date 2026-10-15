# syntax=docker/dockerfile:1.7
ARG SERVICE=alert-worker

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/${SERVICE} ./cmd/${SERVICE}

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/${SERVICE} /${SERVICE}
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/alert-worker"]
