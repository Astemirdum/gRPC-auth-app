FROM golang:1.16

RUN go version
ENV GOPATH=/


COPY ./ ./

RUN go mod download


RUN apt-get update
RUN apt-get -y install postgresql-client

RUN go build -o ./server/cmd/auth-app ./server/cmd/main.go

CMD ["./server/cmd/auth-app"]
