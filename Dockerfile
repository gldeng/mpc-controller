# syntax=docker/dockerfile:1
FROM golang:1.19.0-alpine3.16 as build
WORKDIR /go/src/github.com/avalido/mpc-controller
COPY . .
RUN ls -la
RUN go get -d ./...
RUN apk --update add build-base && cd cmd/mpc-controller && GOOS=linux go build -a -o mpc-controller .

FROM alpine:3.16
WORKDIR /app/
COPY --from=build /go/src/github.com/avalido/mpc-controller/cmd/mpc-controller/mpc-controller ./
CMD ["./mpc-controller"]