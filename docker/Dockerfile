ARG PLATFORM=linux/amd64
FROM --platform=$PLATFORM alpine:latest as initial

RUN mkdir /opt/smr && \
    adduser -D smr-agent && \
    mkdir -p /home/smr-agent/smr/smr && \
    mkdir -p /home/smr-agent/.ssh/simplecontainer && \
    mkdir -p /badger

COPY smr /opt/smr/smr

RUN chmod +x /opt/smr/smr && \
    chown -R smr-agent /opt/smr/smr && \
    chown -R smr-agent /home/smr-agent && \
    chown -R smr-agent /badger && \
    chown -R smr-agent /home/smr-agent/.ssh

FROM scratch

COPY --from=initial / /

USER smr-agent

ENTRYPOINT ["/opt/smr/smr"]
CMD ["start", "--opt"]
EXPOSE 1443