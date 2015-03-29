FROM golang:1.4.2
MAINTAINER anders@janmyr.com

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

CMD ["go-wrapper", "run"]

COPY . /go/src/app

RUN go-wrapper download

WORKDIR /go/src/github.com/coreos/go-etcd
RUN git checkout v0.4.6

WORKDIR /go/src/app
RUN go-wrapper install
