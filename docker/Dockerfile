FROM alpine

WORKDIR /
COPY ./bin/docker_executor /bin

RUN mkdir /cfs

CMD ["docker_executor", "start", "-v"]
