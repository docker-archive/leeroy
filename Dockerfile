FROM debian:jessie
MAINTAINER Jessica Frazelle <jess@docker.com>

ADD https://jesss.s3.amazonaws.com/binaries/go-leeroy /usr/local/bin/go-leeroy

EXPOSE 80

RUN apt-get update && apt-get install -y \
    ca-certificates \
    --no-install-recommends \
    && chmod +x /usr/local/bin/go-leeroy

ENTRYPOINT [ "/usr/local/bin/go-leeroy" ]
