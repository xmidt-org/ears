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
    upx \
    wget \
    unzip

RUN go get github.com/geofffranks/spruce/cmd/spruce && chmod +x /go/bin/spruce
COPY . .
RUN make build

RUN wget -nv https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/download/v0.27.0/otelcontribcol_linux_arm64 \
    && chmod +x otelcontribcol_linux_arm64

# aws-cli v2 does not work on alpine
# see https://stackoverflow.com/questions/61918972/how-to-install-aws-cli-on-alpine

#RUN wget -O "awscliv2.zip" "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" \
#    && unzip awscliv2.zip \
#    && ./aws/install \
#    && sleep 10

# aws-cli v1 needs python

#RUN wget -O "awscli-bundle.zip" "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" \
#    && unzip awscli-bundle.zip \
#    && ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws

RUN apk add --no-cache --no-progress \
        python3 \
        py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install awscli

RUN aws --version && sleep 5

FROM scratch AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/ears /src/NOTICE /src/LICENSE /src/CHANGELOG.md /
COPY --from=builder /src/otelcontribcol_linux_arm64 /

WORKDIR /
ENTRYPOINT [ "/ears" ]
CMD [ "run" ]

HEALTHCHECK \
  --interval=10s --timeout=2s --retries=3 \
  CMD ["/ears", "version"]