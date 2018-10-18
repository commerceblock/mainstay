FROM golang:1.10.3-stretch

COPY . $GOPATH/src/mainstay

RUN set -x \
    && apt update \
    && apt install -y libzmq3-dev \
    && rm -rf /var/lib/apt/lists/*

RUN set -x \
    && cd $GOPATH/src/mainstay \
    && go get ./... \
    && go build \
    && go install

COPY docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["mainstay"]
