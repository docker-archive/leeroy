FROM alpine
MAINTAINER Jessica Frazelle <jess@docker.com>

EXPOSE 80
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
ENV GITHUB_CACHE_PATH /github_cache

RUN	apk update && apk add \
	ca-certificates \
	&& rm -rf /var/cache/apk/*

COPY . /go/src/github.com/docker/leeroy

VOLUME /github_cache

RUN buildDeps=' \
		go \
		git \
		gcc \
		libc-dev \
		libgcc \
	' \
	set -x \
	&& apk update \
	&& apk add $buildDeps \
	&& cd /go/src/github.com/docker/leeroy \
	&& go get -d -v github.com/docker/leeroy \
	&& go build -o /usr/bin/leeroy . \
	&& apk del $buildDeps \
	&& rm -rf /var/cache/apk/* \
	&& rm -rf /go \
	&& echo "Build complete."


ENTRYPOINT [ "leeroy" ]
