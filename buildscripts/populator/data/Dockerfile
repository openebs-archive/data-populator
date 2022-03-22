FROM alpine:3.12

RUN apk add --no-cache bash
RUN apk add --no-cache vim=8.2.4173-r0

COPY bin/data-populator /usr/sbin/data-populator

CMD [ "data-populator" ]
