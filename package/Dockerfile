FROM ubuntu:15.10
RUN apt-get update && \
    apt-get install --no-install-recommends -y curl ca-certificates vim wget bash-completion
RUN curl -L https://github.com/docker/compose/releases/download/1.6.2/docker-compose-Linux-x86_64 > /usr/bin/docker-compose && \
    chmod +x /usr/bin/docker-compose && \
    echo '82129d4e90c7544a52738a570747d653bdb656db  /usr/bin/docker-compose' | sha1sum -c

RUN curl -L https://get.docker.com/builds/Linux/x86_64/docker-1.10.3 > /usr/bin/docker && \
    chmod +x /usr/bin/docker && \
    echo '1886dbb106c5e3b676b03553127a58ce67a9a9ac  /usr/bin/docker' | sha1sum -c

COPY swarm /usr/bin/
COPY swarm-agent /usr/bin/
CMD ["swarm-agent"]
