FROM golang:1.26.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /collector ./collector/cmd/collector


FROM alpine:latest

COPY --from=builder /collector /collector

EXPOSE 50051

CMD ["/collector"]