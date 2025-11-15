FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o authorizer .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/authorizer .

RUN apk --no-cache add ca-certificates tzdata

EXPOSE 9090

ENTRYPOINT ["/app/authorizer"]