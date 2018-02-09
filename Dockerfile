FROM golang:1.9.4-alpine3.6

RUN mkdir -p /code

WORKDIR /code

COPY server.go /code

EXPOSE 8888

RUN apk update && apk add git

RUN go get goji.io && go get gopkg.in/mgo.v2

CMD go run server.go
