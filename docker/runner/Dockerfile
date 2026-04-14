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

COPY scripts/runner/rr-dispatch-ssh /usr/local/bin/rr-dispatch-ssh
COPY scripts/runner/rr-validate-ssh /usr/local/bin/rr-validate-ssh

RUN chmod 0755 /usr/local/bin/rr-dispatch-ssh /usr/local/bin/rr-validate-ssh

USER runner
