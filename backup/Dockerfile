FROM postgres:13.11

RUN apt update; apt install -y wget
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc
RUN mv ./mc /usr/bin
RUN chmod +x /usr/bin/mc

WORKDIR /
COPY ./bin/backup_executor /bin

CMD ["backup_executor", "start"]
