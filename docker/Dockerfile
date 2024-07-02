FROM alpine

RUN apk add --no-cache libc6-compat

RUN mkdir /opt/smr
RUN adduser -D smr-agent

RUN mkdir -p /home/smr-agent/smr/smr
RUN mkdir -p /home/smr-agent/.ssh
RUN mkdir -p /badger

COPY smr/smr /opt/smr-agent/smr
COPY smr/operators/ /home/smr-agent/smr/smr/operators
COPY implementations/ /home/smr-agent/smr/smr/implementations

RUN chmod +x /opt/smr-agent/smr
RUN chown -R smr-agent /opt/smr-agent/smr
RUN chown -R smr-agent /home/smr-agent
RUN chown -R smr-agent /badger
RUN chown -R smr-agent /home/smr-agent/.ssh

USER smr-agent

ENTRYPOINT /opt/smr-agent/smr create smr --daemon --daemon-secured --optmode

EXPOSE 8080