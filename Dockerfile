FROM golang:1.10.3-stretch

COPY . $GOPATH/src/ocean-attestation

RUN set -x \
    && apt update \
    && apt install -y libzmq3-dev \
    && rm -rf /var/lib/apt/lists/*

RUN set -x \
    && cd $GOPATH/src/ocean-attestation \
    && go get ./... \
    && go build \
    && go install

COPY docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["ocean-attestation"]
