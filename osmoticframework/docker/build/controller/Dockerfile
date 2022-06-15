FROM golang:1.17.7-buster

COPY controller /osmoticframework/controller
COPY go.mod /osmoticframework
COPY go.sum /osmoticframework
WORKDIR /osmoticframework

# Thanks Kubernetes team. They screwed up the dependencies
# `go mod` is broken as the names are case-sensitive
# See https://github.com/kubernetes/client-go/issues/741
# They refuse to fix older versions so we're stuck with this mess.
RUN go get -u k8s.io/client-go@v0.18.9 github.com/googleapis/gnostic@v0.3.1
RUN go build -i -o /osmotic_controller osmoticframework/controller
WORKDIR /
ENTRYPOINT /osmotic_controller
