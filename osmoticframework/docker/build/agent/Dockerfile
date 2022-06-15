FROM golang:1.17.7-buster

COPY agent /osmoticframework/agent
COPY go.mod /osmoticframework
COPY go.sum /osmoticframework
WORKDIR /osmoticframework

RUN go build -i -o /osmotic_agent osmoticframework/agent
WORKDIR /
# Before running the container, the host Docker socket must be mounted to the container to access host resources
RUN curl https://get.docker.com | sh
ENTRYPOINT /osmotic_agent

