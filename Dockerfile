# syntax=docker/dockerfile:1
FROM golang:1.18 as build
WORKDIR /go/src/github.com/avalido/
COPY . .
RUN ls -la
RUN cd mpc-controller && go get -d ./...
RUN cd mpc-controller/cmd/mpc-controller && CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o mpc-controller .

FROM alpine:3.16
WORKDIR /app/
COPY --from=build /go/src/github.com/avalido/mpc-controller/cmd/mpc-controller/mpc-controller ./
CMD ["./mpc-controller"]