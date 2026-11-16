FROM golang:1.26.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /gateway ./gateway/cmd/gateway


FROM alpine:latest

COPY --from=builder /gateway /gateway

EXPOSE 8080

CMD ["/gateway"]