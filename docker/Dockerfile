FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest as initial

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN mkdir /opt/smr && \
    adduser -D smr-agent && \
    mkdir -p /home/smr-agent/smr && \
    mkdir -p /home/smr-agent/.ssh/simplecontainer

COPY smr-$TARGETOS-$TARGETARCH/smr /opt/smr/smr

RUN chmod +x /opt/smr/smr && \
    chown -R smr-agent /opt/smr/smr && \
    chown -R smr-agent /home/smr-agent && \
    chown -R smr-agent /home/smr-agent/.ssh

RUN apk add busybox-extras

FROM scratch

COPY --from=initial / /

USER smr-agent

ENTRYPOINT ["/opt/smr/smr"]
CMD ["start", "--opt"]
EXPOSE 1443