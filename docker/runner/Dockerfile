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
COPY scripts/runner/start-runner /usr/local/bin/start-runner

RUN chmod 0755 /usr/local/bin/rr-exec /usr/local/bin/start-runner \
    && mkdir -p /home/runner/.runner-state \
    && chown runner:runner /home/runner/.runner-state

USER runner
