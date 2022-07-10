FROM golang:1.18

ENV GOPROXY="https://goproxy.io"
WORKDIR /usr/src/app

EXPOSE 8206
CMD ["ftc-api-v6", "-production=false", "-livemode=false"]

COPY . .
RUN go mod download && go mod verify
RUN go build -o /usr/local/bin/ftc-api-v6 -tags production .
