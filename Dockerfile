FROM golang:1.13-rc-alpine AS builder

ENV GO111MODULE=on

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY go.mod .
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./tmp/app

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/app .
CMD ["./tmp/app"] 