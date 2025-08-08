FROM scratch 

COPY bin/linux/ino /bin/ino
COPY /migrations /migrations
ENV INO_MIGRATIONS_PATH=file:///migrations

ENTRYPOINT ["ino"]
