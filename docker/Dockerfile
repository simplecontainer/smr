FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest as initial

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN mkdir /opt/smr && \
    adduser -D node && \
    mkdir -p /home/node/smr && \
    mkdir -p /home/node/.ssh/simplecontainer

COPY smr-$TARGETOS-$TARGETARCH/smr /opt/smr/smr

RUN chmod +x /opt/smr/smr && \
    chown -R node /opt/smr/smr && \
    chown -R node /home/node && \
    chown -R node /home/node/.ssh && \
    touch /home/node/.ssh/known_hosts

RUN apk add busybox-extras

FROM scratch

COPY --from=initial / /

USER node

ENTRYPOINT ["/opt/smr/smr"]
CMD ["start"]
EXPOSE 1443