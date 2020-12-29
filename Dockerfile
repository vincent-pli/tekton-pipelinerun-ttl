FROM ubuntu

WORKDIR /

COPY _output/bin/taskrun-watcher /usr/local/bin

ENTRYPOINT []
CMD ["taskrun-watcher"]
