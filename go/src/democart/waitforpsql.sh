#!/bin/sh
# waitforpsql.sh

set -e
  
cmd="$@"
  
# TODO(sam): pull this in from docker environent
until PGPASSWORD="democartpass" psql -h "db" -U "democart" -d "democart" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done
  
>&2 echo "Postgres is up - executing command"
exec $cmd
