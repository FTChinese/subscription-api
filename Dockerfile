FROM golang:1.18 AS builder

ENV GOPROXY="https://goproxy.io"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /app/out/ftc-api-v6 -tags production .
RUN chmod +x /app/out/ftc-api-v6

FROM ubuntu
EXPOSE 8206
WORKDIR /app
CMD ["./ftc-api-v6", "-production=false", "-livemode=false"]
COPY --from=builder /app/out/ftc-api-v6 .

