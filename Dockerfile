FROM alpine

RUN apk add --no-cache --update ca-certificates

ENV GOOGLE_APPLICATION_CREDENTIALS=/conf/hgnet-tea.json

WORKDIR /app

COPY /go/bin/app/teadb /app/

ENTRYPOINT ["/app/teadb"]
