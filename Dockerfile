FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /fotoferry .

FROM debian:bookworm-slim
ARG S6_OVERLAY_VERSION=3.2.0.0
ARG TARGETARCH

RUN apt-get update && \
    apt-get install -y --no-install-recommends cron gettext-base xz-utils ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

ADD https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz /tmp/s6-noarch.tar.xz
RUN case "${TARGETARCH}" in \
        amd64) S6_ARCH=x86_64 ;; \
        arm64) S6_ARCH=aarch64 ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    curl -fsSL "https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-${S6_ARCH}.tar.xz" -o /tmp/s6-arch.tar.xz && \
    tar -C / -Jxpf /tmp/s6-noarch.tar.xz && \
    tar -C / -Jxpf /tmp/s6-arch.tar.xz && \
    rm /tmp/s6-noarch.tar.xz /tmp/s6-arch.tar.xz

COPY --from=build /fotoferry /usr/local/bin/fotoferry
COPY docker/crontab.template /app/crontab.template
COPY docker/etc/s6-overlay /etc/s6-overlay
RUN find /etc/s6-overlay -type f \( -name run -o -name up -o -name finish \) -exec chmod +x {} \; && \
    find /etc/s6-overlay/scripts -type f -exec chmod +x {} \;

ENTRYPOINT ["/init"]
