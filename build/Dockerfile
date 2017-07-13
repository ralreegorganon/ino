FROM alpine:3.6

RUN apk --no-cache add ca-certificates

ADD bin/linux/ino /bin/ino

ENTRYPOINT ["ino"]