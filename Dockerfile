FROM alpine

RUN apk add --no-cache --update ca-certificates

ENV GOOGLE_APPLICATION_CREDENTIALS=/conf/hgnet-tea.json

COPY teadbd/teadbd .

ENTRYPOINT ["./teadb"]
