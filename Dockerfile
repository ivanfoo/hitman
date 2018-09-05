FROM golang:1.10-alpine3.7 as build

WORKDIR /go/src/hitman
COPY . .

RUN apk add --no-cache git
RUN go get -v .
RUN go install -v .

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/hitman /
CMD ["/hitman"]
