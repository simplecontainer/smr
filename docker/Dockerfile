FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN adduser -D node && \
    mkdir -p /opt/smr /home/node/smr /home/node/.ssh && \
    chown -R node /home/node /opt/smr && \
    touch /home/node/.ssh/known_hosts

COPY smr-${TARGETOS}-${TARGETARCH}/smr /opt/smr/smr
RUN chmod +x /opt/smr/smr && \
    # Install necessary packages
    apk add --no-cache busybox-extras ca-certificates tzdata

FROM alpine:latest

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /home/node /home/node
COPY --from=builder /opt/smr /opt/smr
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /home/node/smr

USER node

EXPOSE 1443

ENTRYPOINT ["/opt/smr/smr"]
CMD ["start"]