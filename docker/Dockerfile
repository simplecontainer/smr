FROM alpine:latest as initial

ENV GLIBC_REPO=https://github.com/sgerrand/alpine-pkg-glibc
ENV GLIBC_VERSION=2.34-r0

RUN set -ex && \
    apk --update add libstdc++ curl ca-certificates && \
    for pkg in glibc-${GLIBC_VERSION} glibc-bin-${GLIBC_VERSION}; \
        do curl -sSL ${GLIBC_REPO}/releases/download/${GLIBC_VERSION}/${pkg}.apk -o /tmp/${pkg}.apk; done && \
    apk add --allow-untrusted --force-overwrite /tmp/*.apk && \
    rm -v /tmp/*.apk && \
    /usr/glibc-compat/sbin/ldconfig /lib /usr/glibc-compat/lib

RUN mkdir /opt/smr && \
    adduser -D smr-agent && \
    mkdir -p /home/smr-agent/smr/smr && \
    mkdir -p /home/smr-agent/.ssh && \
    mkdir -p /badger

COPY smr /opt/smr-agent/smr
COPY operators/**/*.so /home/smr-agent/smr/smr/operators/
COPY implementations/**/*.so /home/smr-agent/smr/smr/implementations/

RUN chmod +x /opt/smr-agent/smr && \
    chown -R smr-agent /opt/smr-agent/smr && \
    chown -R smr-agent /home/smr-agent && \
    chown -R smr-agent /badger && \
    chown -R smr-agent /home/smr-agent/.ssh

FROM alpine:latest

COPY --from=initial / /

USER smr-agent

ENTRYPOINT /opt/smr-agent/smr create smr --daemon --daemon-secured --optmode
EXPOSE 8080

