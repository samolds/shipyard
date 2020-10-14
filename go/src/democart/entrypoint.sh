#!/bin/sh
# entrypoint.sh

set -e

cmd="$@"

if [ "$DATABASE_DRIVER" = "postgres" ]
then
  echo "Waiting for Postgres at ${DATABASE_HOST}:${DATABASE_PORT}..."

  while ! nc -z ${DATABASE_HOST} ${DATABASE_PORT}; do
    sleep 0.1
  done

  echo "Postgres started"
fi

exec $cmd
