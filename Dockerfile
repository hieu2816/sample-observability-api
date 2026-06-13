FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X sample-observability-api/internal/buildinfo.version=${VERSION} -X sample-observability-api/internal/buildinfo.commit=${COMMIT} -X sample-observability-api/internal/buildinfo.buildTime=${BUILD_TIME}" \
    -o /out/sample-observability-api ./cmd/api

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

EXPOSE 8080
COPY --from=build /out/sample-observability-api /sample-observability-api
ENTRYPOINT ["/sample-observability-api"]
