FROM golang:1.21 as build-env
RUN go install github.com/mhmdiaa/chronos@latest

FROM alpine:3.20
RUN apk add --no-cache bind-tools ca-certificates libc6-compat
COPY --from=build-env /go/bin/chronos /usr/local/bin/chronos
ENTRYPOINT ["chronos"]
