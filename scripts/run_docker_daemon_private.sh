docker run -d \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -v /tmp:/tmp -p 127.0.0.1:8080:8080 \
      --name smr-agent \
      --dns 127.0.0.1 \
      smr:0.0.1