FROM golang:1.13-stretch AS builder

ENV GO111MODULE=on

# setup the container
RUN apt-get update && apt-get install -y sqlite3

# setup
WORKDIR /usr/local/src/democart
COPY go.mod go.sum Makefile ./
RUN make setup
COPY . .
RUN make test server


FROM debian:stretch AS runner

# TODO(sam): create a non root user and give them permissions
RUN apt-get update && apt-get install -y ca-certificates
RUN mkdir -p /var/democart
RUN mkdir -p /etc/democart
COPY config.hcl /etc/democart/config.hcl
# TODO(sam): how to use this without needing to install psql in the image?
#COPY waitforpsql.sh /usr/local/bin/waitforpsql.sh
COPY --from=builder /usr/local/src/democart/democart /usr/local/bin/democart

#CMD [ "/usr/local/bin/waitforpsql.sh", \
CMD [ "/usr/local/bin/democart", \
  "--config", "/etc/democart/config.hcl" ]

EXPOSE 8080
