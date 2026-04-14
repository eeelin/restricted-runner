FROM ghcr.io/actions/actions-runner:latest

USER root

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        openssh-client \
        jq \
        git \
        bash \
    && rm -rf /var/lib/apt/lists/*

COPY scripts/runner/rr-exec /usr/local/bin/rr-exec

RUN chmod 0755 /usr/local/bin/rr-exec

USER runner
