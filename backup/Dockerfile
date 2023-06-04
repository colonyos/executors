FROM postgres:13.11

WORKDIR /
COPY ./bin/backup_executor /bin

CMD ["backup_executor", "start"]
