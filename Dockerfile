FROM ubuntu

ENV DEBIAN_FRONTEND noninteractive

RUN apt update

# configure locale

ENV TZ=America/Chicago
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US.UTF-8
ENV LC_ALL en_US.UTF-8

RUN set -xe \
    && apt-get install -y apt-utils tzdata locales
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone
RUN set -xe && \
    dpkg-reconfigure --frontend=noninteractive tzdata && \
    echo en_US.UTF-8 UTF-8 > /etc/locale.gen && \
    echo 'LANG="en_US.UTF-8"'>/etc/default/locale && \
    dpkg-reconfigure --frontend=noninteractive locales && \
    update-locale LANG=en_US.UTF-8

# https://wiki.ubuntu.com/Minimal
RUN echo "y\ny" | unminimize

RUN apt install -y \
    build-essential curl ca-certificates unzip sudo man jq gettext-base \
    iputils-ping \
    golang \
    docker.io \
    neovim \
    ;

ENV TERRAFORM_VERSION 0.12.24
RUN mkdir -p /tmp/terraform && cd /tmp/terraform \
    && curl -o terraform.zip -L https://releases.hashicorp.com/terraform/"$TERRAFORM_VERSION"/terraform_"$TERRAFORM_VERSION"_linux_amd64.zip \
    && unzip terraform.zip \
    && mv /tmp/terraform/terraform /usr/local/bin/terraform \
    && rm -rf /tmp/terraform

COPY . /bla
RUN cd /bla && go build -o /toolbox-init init/main.go
RUN cd /bla && go build -o /toolbox-run run/main.go
RUN cd /bla && env GOOS=darwin GOARCH=amd64 go build -o /toolbox-stub stub/main.go

COPY root /

ENV DEBIAN_FRONTEND=
ENTRYPOINT ["/toolbox-init"]
CMD ["/bin/bash"]
