FROM golang:1.17.1-alpine as build-env
RUN go get -v github.com/mhmdiaa/chronos

FROM alpine:3.14
RUN apk add --no-cache bind-tools ca-certificates
COPY --from=build-env /go/bin/chronos /usr/local/bin/chronos
ENTRYPOINT ["chronos"]
