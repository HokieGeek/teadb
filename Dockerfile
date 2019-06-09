### Build the code
FROM golang:1.12-alpine

ENV GO111MODULE on
ENV GOOS linux
ENV GOARCH amd64

RUN apk add --update git

WORKDIR /go/src/gitlab.com/hokiegeek.net/teadb

COPY . .

RUN go install -v -ldflags="-w -s" ./...

### Package it up
FROM alpine

VOLUME /conf

EXPOSE 80

ENV GOOGLE_APPLICATION_CREDENTIALS=/conf/hgnet-teadb.json

RUN apk add --no-cache --update ca-certificates

# RUN addgroup -S gouser && adduser -S -G gouser gouser

# USER gouser

COPY --from=0 /go/bin/teadbd .

ENTRYPOINT ["./teadbd"]
