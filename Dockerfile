FROM golang:1.13-rc-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

ENV GO111MODULE=on

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY go.mod .
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./tmp/app

FROM alpine:latest  
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /go/src/app/tmp/app .
CMD ["./app"] 