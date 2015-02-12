FROM debian:jessie
MAINTAINER Jessica Frazelle <jess@docker.com>

ADD https://jesss.s3.amazonaws.com/binaries/leeroy /usr/local/bin/leeroy

EXPOSE 80

RUN apt-get update && apt-get install -y \
    ca-certificates \
    --no-install-recommends \
    && chmod +x /usr/local/bin/leeroy

ENTRYPOINT [ "/usr/local/bin/leeroy" ]
