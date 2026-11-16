FROM golang:1.26.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /gateway ./cmd/gateway


FROM alpine:3.23

RUN apk add --no-cache ca-certificates

COPY --from=builder /gateway /gateway

EXPOSE 8080

CMD ["/gateway"]
