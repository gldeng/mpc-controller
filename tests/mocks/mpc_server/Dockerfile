# syntax=docker/dockerfile:1
FROM golang:1.19.0-alpine3.16 as build
WORKDIR /go/src/github.com/avalido/mpc-controller
COPY . .
RUN go get -d -v ./...
RUN apk --update add build-base && cd tests/mocks/mpc_server && GOOS=linux go build -a -o mpc-server-mock .

FROM alpine:3.16
WORKDIR /app/
COPY --from=build /go/src/github.com/avalido/mpc-controller/tests/mocks/mpc_server/mpc-server-mock ./
CMD ["./mpc-server-mock"]