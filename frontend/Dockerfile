FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY servefrontend.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o /frontend-server servefrontend.go

FROM alpine:3.22

WORKDIR /srv

RUN apk add --no-cache ca-certificates

COPY --from=builder /frontend-server /frontend-server
COPY frontend ./frontend

EXPOSE 5500

CMD ["/frontend-server", "-addr", ":5500", "-dir", "frontend"]
