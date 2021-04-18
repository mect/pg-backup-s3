ARG PG
FROM golang:1.15-alpine as build

COPY ./ /go/src/github.com/mect/pg-backup-s3

WORKDIR /go/src/github.com/mect/pg-backup-s3
RUN go build -o pg-backup-s3 ./cmd/pg-backup-s3

ARG PG
FROM postgres:${PG}-alpine

COPY --from=build /go/src/github.com/mect/pg-backup-s3/pg-backup-s3 /usr/local/bin/pg-backup-s3

ENTRYPOINT ["pg-backup-s3"]
