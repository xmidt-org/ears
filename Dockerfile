FROM docker.io/library/golang:1.19-alpine as builder

WORKDIR /src

ARG VERSION
ARG GITCOMMIT
ARG BUILDTIME

RUN apk add --no-cache --no-progress \
    ca-certificates \
    make \
    curl \
    git \
    openssh \
    gcc \
    libc-dev \
    upx \
    wget

RUN mkdir -p /go/bin && \
    curl -o /go/bin/spruce https://github.com/geofffranks/spruce/releases/download/v1.29.0/spruce-linux-amd64 && \
    chmod +x /go/bin/spruce
COPY . .
RUN make build

RUN wget -nv https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/download/v0.27.0/otelcontribcol_linux_amd64 \
    && chmod +x otelcontribcol_linux_amd64

FROM alpine:latest AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/ears /src/NOTICE /src/LICENSE /src/CHANGELOG.md /
COPY --from=builder /src/otelcontribcol_linux_amd64 /otelcontribcol

# Install bash in the runtime stage
RUN apk add --no-cache bash

WORKDIR /
ENTRYPOINT [ "/ears" ]
CMD [ "run" ]

HEALTHCHECK \
  --interval=10s --timeout=2s --retries=3 \
  CMD ["/ears", "version"]
