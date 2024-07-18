FROM colonyos/colonies:v1.8.7

WORKDIR /
COPY ./bin/docker_executor /bin
RUN mkdir /cfs
CMD ["docker_executor", "start", "-v"]
