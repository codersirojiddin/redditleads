# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /out/api ./apps/api
RUN go build -o /out/worker ./apps/worker

FROM alpine:3.19 AS api
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api ./api
COPY migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./api"]

FROM alpine:3.19 AS worker
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/worker ./worker
COPY migrations ./migrations
ENTRYPOINT ["./worker"]
