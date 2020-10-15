### builder image
FROM golang:1.15-buster AS builder

# install dependencies
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends sqlite3 && \
    apt-get clean

# add and install requirements
ENV GO111MODULE=on

# setup, test, and build bin from src
ENV HOME=/usr/local/src/democart
WORKDIR $HOME
COPY go.mod go.sum Makefile $HOME/
RUN make setup-no-dbx
COPY . $HOME/

# version arg supplied docker-compose and used in "make server"
ARG VERSION
RUN make test-no-dbx server-no-dbx




### runner image
FROM debian:buster-slim AS runner

# install dependencies
# netcat-openbsd is needed by nc in entrypoint.sh to detect postgres boot
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
      netcat-openbsd ca-certificates && \
    apt-get clean

# name app user
ENV APP_USER=appuser

# add $APP_USER
RUN addgroup --system $APP_USER && \
  adduser --system --group $APP_USER

# switch to non-root $APP_USER
USER $APP_USER

# copy in binaries and config files
COPY --from=builder /usr/local/src/democart/democart /usr/local/bin/democart
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY config.hcl /etc/democart/dockerfile_config.hcl

CMD ["/usr/local/bin/entrypoint.sh", \
  "/usr/local/bin/democart", \
  "--configs=/etc/democart/dockerfile_config.hcl"]

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082
