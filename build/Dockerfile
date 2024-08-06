FROM alpine:3.20.2

RUN apk --no-cache add ca-certificates

ADD bin/linux/ino /bin/ino
ADD /migrations /migrations
ENV INO_MIGRATIONS_PATH=file:///migrations

ENTRYPOINT ["ino"]