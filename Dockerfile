FROM golang:1.10.1 AS builder
WORKDIR /go/src/github.com/keadwen/msc_project
COPY . /go/src/github.com/keadwen/msc_project

RUN go get github.com/golang/protobuf/proto
RUN go get -v gonum.org/v1/plot
RUN go build -o /bin/msc_project

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN go version
ENTRYPOINT ["/bin/bash"] 
