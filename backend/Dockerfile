FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build \
    -o /server \
    ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates libgcc

COPY --from=builder /server /server

EXPOSE 8080

CMD ["/server"]
