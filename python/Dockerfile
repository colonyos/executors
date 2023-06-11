FROM ubuntu:22.04 

ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Europe/Stockholm

RUN apt-get update; apt-get -y install wget ca-certificates tzdata python3 python3-pip 
RUN pip3 install pycolonies

WORKDIR /
COPY ./python_executor.py / 
COPY ./libcryptolib.so /usr/local/lib 

CMD ["python3", "/python_executor.py"]
