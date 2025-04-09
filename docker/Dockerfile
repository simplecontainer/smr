FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest as initial

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN adduser -D node && \
    mkdir -p /opt/smr /home/node/smr /home/node/.ssh/simplecontainer && \
    chown -R node:node /home/node /opt/smr && \
    touch /home/node/.ssh/known_hosts

COPY smr-$TARGETOS-$TARGETARCH/smr /opt/smr/smr
RUN chmod +x /opt/smr/smr

RUN apk add --no-cache busybox-extras

FROM scratch

COPY --from=initial / /

# Set user to 'node' and specify entrypoint and command
USER node
ENTRYPOINT ["/opt/smr/smr"]
CMD ["start"]

# Expose port 1443
EXPOSE 1443
