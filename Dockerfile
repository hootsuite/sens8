FROM alpine
MAINTAINER Mark Eijsermans <mark.eijsermans@hootsuite.com>

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ADD sens8 /usr/local/bin/sens8

ENTRYPOINT /usr/local/bin/sens8
CMD ["-config-file=/etc/sensu/config/config.json", "-logtostderr=true"]

