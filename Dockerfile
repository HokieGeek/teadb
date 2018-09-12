### Build the code
FROM golang:1.10-alpine

RUN apk add --update git

COPY . /go/src/gitlab.com/hokiegeek.net/teadb

WORKDIR /go/src/gitlab.com/hokiegeek.net/teadb

RUN go get -d -v ./...
RUN go install -v ./...
RUN go test -v ./...

### Package it up
FROM alpine

EXPOSE 80
EXPOSE 443

RUN apk add --no-cache --update ca-certificates

ENV GOOGLE_APPLICATION_CREDENTIALS=/conf/hgnet-teadb.json

COPY --from=0 /go/bin/teadbd .
# COPY teadbd/teadbd .

ENTRYPOINT ["./teadbd"]
