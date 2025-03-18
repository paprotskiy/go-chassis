FROM alpine:latest as alpine-extended
RUN apk add gcompat

FROM golang:latest AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./src ./src
ARG VERSION
RUN cd ./src/cmd/app && go build -ldflags "-s -w -X main.version=${VERSION}" -o exec

FROM alpine-extended

RUN addgroup -g 1000 appgroup && \
  adduser -D -u 1000 -G appgroup -s /bin/sh appuser
USER appuser

WORKDIR /app

COPY ./src/internal/adapters/storage/migrations/ ./src/internal/adapters/storage/migrations/
COPY --from=builder /build/src/cmd/app/exec ./exec
CMD ["./exec"]
