FROM golang:1.10-alpine as build
RUN apk add --no-cache git make
RUN mkdir -p /go/src/github.com/segmentio/tracking-api-chaos/vendor
COPY ./vendor/vendor.json /go/src/github.com/segmentio/tracking-api-chaos/vendor/vendor.json
WORKDIR /go/src/github.com/segmentio/tracking-api-chaos
# these are split out so we can cache the vendor step
RUN make vendor
COPY . /go/src/github.com/segmentio/tracking-api-chaos
RUN make

FROM alpine:3.7
EXPOSE 8080
COPY --from=build /go/src/github.com/segmentio/tracking-api-chaos/tracking-api-chaos /
RUN chmod +x /tracking-api-chaos
ENTRYPOINT ["/tracking-api-chaos"]
