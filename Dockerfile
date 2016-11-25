FROM golang:alpine
MAINTAINER sklirg

ENV APP=/go/src/github.com/sklirg/wcgstats-scraper/

RUN apk add --no-cache --update git

RUN mkdir -p $APP
WORKDIR $APP

COPY . .
RUN go get
RUN go build

CMD ["./wcgstats-scraper"]
