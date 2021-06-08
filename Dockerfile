FROM docker.io/library/golang:1.15-alpine as builder

MAINTAINER Jack Murdock <jack_murdock@comcast.com>

WORKDIR /src

ARG VERSION
ARG GITCOMMIT
ARG BUILDTIME


RUN apk add --no-cache --no-progress \
    ca-certificates \
    make \
    git \
    openssh \
    gcc \
    libc-dev \
    upx

RUN go get github.com/geofffranks/spruce/cmd/spruce && chmod +x /go/bin/spruce
COPY . .
RUN make build

FROM scratch AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/ears /src/NOTICE /src/LICENSE /src/CHANGELOG.md /

WORKDIR /
ENTRYPOINT [ "/ears" ]
CMD [ "run" ]

HEALTHCHECK \
  --interval=10s --timeout=2s --retries=3 \
  CMD ["/ears", "version"]