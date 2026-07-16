FROM golang:1.23

# Client mysql nécessaire pour appliquer schema.sql/seeds.sql automatiquement
RUN apt-get update && apt-get install -y --no-install-recommends default-mysql-client \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -g 1000 go && useradd -d /home/go -g 1000 -m go

# GOPATH/cache modules en dehors de /home/go pour survivre au bind mount de dev
ENV GOPATH=/go
ENV GOMODCACHE=/go/pkg/mod
RUN mkdir -p /go && chown -R go:go /go

USER go
WORKDIR /home/go

CMD ["sh", "docker-entrypoint.sh"]
