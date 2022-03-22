FROM alpine:3.12

RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.4173-r0
RUN apk add --no-cache rsync==3.1.3-r3

# entrypoint script
COPY buildscripts/rsync/daemon/entrypoint.sh /usr/sbin/entrypoint.sh
RUN chmod +x /usr/sbin/entrypoint.sh
RUN mkdir -p /entrypoint.d

ENTRYPOINT [ "entrypoint.sh" ]

EXPOSE 873
CMD [ "/usr/bin/rsync", "--no-detach", "--daemon", "--log-file=/dev/stdout" ]
