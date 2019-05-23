FROM golang:1.12-alpine

ENV GOPATH /go
ENV GO111MODULE on

RUN apk add --no-cache git
WORKDIR /flakbase
ADD . /flakbase
RUN go install

FROM alpine:3.9

EXPOSE 9527
COPY --from=0 /go/bin/flakbase /usr/bin/flakbase

CMD flakbase