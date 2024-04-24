FROM alpine

RUN apk add --no-cache libc6-compat

RUN mkdir /opt/smr
RUN adduser -D smr-agent

RUN mkdir -p /home/smr-agent/smr/smr
RUN mkdir -p /badger

COPY smr /opt/smr-agent/smr
COPY operators/ /home/smr-agent/smr/smr/operators
COPY implementations/ /home/smr-agent/smr/smr/implementations

RUN chmod +x /opt/smr-agent/smr
RUN chown -R smr-agent /opt/smr-agent/smr
RUN chown -R smr-agent /home/smr-agent
RUN chown -R smr-agent /badger

USER smr-agent

ENTRYPOINT /opt/smr-agent/smr create smr --daemon --optmode

EXPOSE 8080