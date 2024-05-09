#Create directory for persisting ca certificate, server certificate and client certificate
mkdir ~/.smr-pki

#Run docker: mount docker.sock from the host, tmp for mounting resources and point dns to itself since we are handling dns
docker run -d \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /tmp:/tmp \
       -v ~/.smr-pki/:/home/smr-agent/.ssh \
       -p 0.0.0.0:1443:1443 \
       --name smr-agent \
       --dns 127.0.0.1 \
       smr:0.0.1