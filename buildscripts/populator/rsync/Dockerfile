FROM alpine:3.12

RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.4173-r0

COPY bin/rsync-populator /usr/sbin/rsync-populator

ENTRYPOINT [ "rsync-populator" ]
