ARG ARCH
ARG GOARCH
FROM golang:1.15-alpine as build

COPY ./ /go/src/github.com/mect/pg-backup-s3

WORKDIR /go/src/github.com/mect/pg-backup-s3
ENV GOARM=7
ENV GOARCH=$GOARCH
RUN go build -o pg-backup-s3 ./cmd/pg-backup-s3

ARG ARCH
FROM $ARCH/postgres:12-alpine

COPY --from=build /go/src/github.com/mect/pg-backup-s3/pg-backup-s3 /usr/local/bin/pg-backup-s3

ENTRYPOINT ["pg-backup-s3"]
