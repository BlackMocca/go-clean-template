FROM golang:1.26 as build

RUN mkdir -p /go/src/application
WORKDIR /go/src/application

ENV GO111MODULE=on
ENV ADDR=0.0.0.0
ENV TZ=Asia/Bangkok

COPY go.mod .
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/src/application/build/app main.go

FROM alpine:3.22
RUN apk --no-cache add ca-certificates tzdata

ENV ADDR=0.0.0.0
ENV TZ=Asia/Bangkok

WORKDIR /usr/app

COPY --from=build /go/src/application/build/app .
COPY --from=build /go/src/application/assets assets
COPY --from=build /go/src/application/migrations migrations

EXPOSE 3000
EXPOSE 3001
EXPOSE 3100

CMD ["./app"]