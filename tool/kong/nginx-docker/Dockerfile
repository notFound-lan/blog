FROM nginx:1.19.6

SHELL [ "/bin/bash", "-c" ]

COPY sources.list /etc/apt/sources.list

RUN apt update -y \
  && apt install -y \
    vim \
    curl \
    procps \
    tree \
    net-tools \
    apache2-utils \
    netcat